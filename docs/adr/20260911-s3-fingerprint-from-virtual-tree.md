---
title: "20260911 - Fingerprint S3 buckets from a virtual tree; object keys never become local paths"
description: "Download each object to an anonymous temp file, hash it, delete it, and compute the directory fingerprint from (key, sha256) pairs so that no S3 key is ever used as a filename"
status: "Accepted"
date: "2026-09-11"
---

# 20260911 - Fingerprint S3 buckets from a virtual tree; object keys never become local paths

## Overview

`kosli snapshot s3` stops recreating the bucket as a directory on the operator's machine. Each object is downloaded to an anonymous temp file, hashed, and deleted. The fingerprint is computed by `digest.VirtualDirSha256` from `(key, sha256)` pairs, which reproduces `digest.DirSha256` byte for byte. Content mode and the metadata mode of #1069 become one pipeline that differs only in where each object's sha256 comes from.

## Context

Today `getS3DataFromClient` writes every object to `filepath.Join(tempDir, key)` and fingerprints the resulting tree with `DirSha256`. The object key, which anyone with write access to the bucket controls, therefore names a file on the operator's filesystem.

PR #1155 fenced that key after a privately reported traversal: it rejects `..` segments, opens each destination with `O_EXCL` so two keys cannot share a file, and turns `ENOTDIR` into a readable error. It closes the reported hole, but the key still shapes a path, so a class of problems remains:

- Two directories differing only in case merge on macOS and Windows. Unicode normalisation on macOS does the same. Both are fingerprint collisions between distinct bucket contents.
- Legitimate keys fail: `...` and `.. ` directories everywhere, and on Windows also `CON`, any colon, and a leading backslash, purely because Windows would misread the name on disk.
- A key component over 255 bytes fails with a bare `ENAMETOOLONG`.
- The fingerprint depends on the platform the snapshot runs on. A bucket with `A/x` and `a/y`, or with `d\e.txt`, fingerprints differently on Linux than on macOS or Windows.
- Around sixty lines of production code and most of the #1155 test table reason about Windows and macOS filesystem semantics that CI never executes.

Separately, #1069 (`--fingerprint-source metadata`) built `VirtualDirSha256` to reproduce `DirSha256` without a filesystem, and its key rule (`validateVirtualPath`, strict `path.Clean` equality) disagrees with #1155's on-disk rule in both directions: `a//b`, `./c.txt` and `/lead.txt` are accepted on disk and rejected virtually; `...` is rejected on disk and accepted virtually. A bucket that snapshots today would break when the default flips.

The fingerprint format itself is fixed. An S3 snapshot must match the fingerprint of the deployed directory attested earlier with `kosli attest artifact --artifact-type dir`, and every existing environment snapshot on the server was computed as `DirSha256` of the tree the key layout produced. So the change must preserve, for every bucket that snapshots successfully today, the identical fingerprint and artifact name.

## Decision

1. **The key never touches the filesystem.** Each object is downloaded to `os.CreateTemp(tempDir, "object-*")`, hashed with `FileSha256`, closed and removed. Peak disk drops from the bucket size to the in-flight objects. The transfer manager keeps working unchanged because `*os.File` is an `io.WriterAt`, and so does `FakeS3Client`.

2. **The key becomes a virtual path by one rule, shared by both modes.** A key holding a `..` segment is rejected first, on the raw key, so `a/../b` cannot fold silently onto `b`. The rest is `path.Clean(strings.TrimLeft(key, "/"))`, rejecting a result of `.`. The fold is exactly what `filepath.Join` produced on Linux, so `a//b`, `./c.txt` and `/lead.txt` land where they do today. Everything else, including `...`, `.. `, `CON`, colons and backslashes, is an ordinary name because nothing is created under it. None of this is for safety. `DirSha256` walks real directories, which can never hold an entry named `.`, `..` or the empty string, so the rule keeps virtual fingerprints inside the space an attested directory can match, and preserves the fingerprints existing buckets already have.

3. **Collisions are data errors that name every key involved.** Two keys resolving to one path, and a key under a prefix that is also an object, cannot be represented as a tree. They are detected before the digest package sees them and reported together, bounded by `maxReportedS3KeyProblems` in both directions, at most ten problem lines each naming at most ten keys, with the existing advice to use `--exclude-regex` or narrow the include filter.

4. **A root `.kosli_ignore` is honoured virtually.** Its rules are parsed by `digest.ParseIgnoreRules`, the same reading `DirSha256` gives the file, and resolved by `digest/virtualglob.go`, which reproduces `filepathx.Glob`, `filepath.Glob` and `filepath.Walk` step for step over the virtual tree rather than reimplementing what the globs appear to mean. That is what keeps their quirks identical: a literal `**/x` finds a root `x` spelled with a double slash, which the walk's cleaned paths never equal, so a root file `x` survives while a root directory `x` keeps its name and loses its contents; `**/*.log` is rebuilt cleaned and matches outright; and excluding `logs/*` leaves an empty directory whose name is still hashed. Exclusion therefore runs inside the tree walk, not by filtering the file list. The ignore file can never exclude itself, as in `DirSha256`. `digest.FilesNeedingContent` shares that walk so excluded objects are not downloaded at all, and `VirtualDirSha256` refuses a tree that needs a digest it was not given, so a skipped download can never leak into a fingerprint. Equivalence with `DirSha256` on a materialised tree is asserted for every rule shape in `TestVirtualIgnoreTestSuite`.

5. **Downloads may run in parallel**, behind a fixed worker pool and a bytes-in-flight budget derived from listing sizes, with results written by index for determinism and a cancellable context so the first transport error stops in-flight multipart downloads. This is a performance property, not a safety one, and was delivered separately from the change this record describes; see #1167. The defaults are eight objects in flight within 512 MB of listed bytes, tunable with `--download-concurrency` and `--download-budget`, so peak temp disk is the budget rather than one object.

6. **The switch is pinned, not argued.** `TestPinnedFingerprints` in `internal/aws` holds fingerprints of representative fake buckets recorded against the key-layout implementation before it was replaced: unusual key shapes, a `.` sorting before a `/`, nested prefixes with folder markers, and a single object with its basename as artifact name. It was green before the switch and must stay green after it, alongside `TestGetS3DataFromClientKeepsTodaysLayoutForUnusualKeys` from #1155.

## Guarantees

The attest side is unchanged: `kosli attest artifact --artifact-type dir` still runs `DirSha256` on a real directory, and that defines the format. The snapshot side re-derives the same value from the bucket. Three guarantees follow, each with the test that holds it:

- **Snapshot equals attestation.** `VirtualDirSha256` of the bucket's `(path, sha256)` pairs equals `DirSha256` of the directory those objects were uploaded from, and a single object equals `FileSha256` plus its basename. Held by the materialise-then-compare tests in `internal/digest` (`TestVirtualDirTestSuite`, `TestVirtualIgnoreTestSuite`) and by `TestMatchesAttestedDirectory` in `internal/aws`, which hashes a directory, uploads its files to the fake bucket and snapshots it.
- **Snapshot equals its own history.** Every bucket that snapshotted successfully under the key-layout implementation keeps its fingerprint and artifact name, because the fold rule is what `filepath.Join` did. Held by `TestPinnedFingerprints` and `TestGetS3DataFromClientKeepsTodaysLayoutForUnusualKeys`.
- **No object is lost silently.** Every key S3 accepts is representable, whatever the operating system, because the key never becomes a path. Rejections are of two kinds, and each fails the snapshot rather than shorten the manifest. Keys that cannot form a directory tree at all: a `..` segment, two keys folding onto one path, and an object that is also a prefix. S3 permits those, no filesystem does, so no attested directory could match them, and the error names every key involved. And listings the store returned malformed: an entry with no key, a key listed more than once, and no modification time on any matching object. Real S3 produces none of these; an S3-compatible store might, and each is an error rather than a guess. A path deeper than `globSeparatorsLimit` segments is also rejected, which no S3 key can reach, so the exported digest functions carry the bound for other callers. Held by `TestEveryS3KeyIsRepresentable`, which generates keys over S3's full character set, and `TestDownloadsExactlyTheContributingObjects`, which asserts the downloaded set equals the listed objects minus markers and exclusions.

## Alternatives considered

- **Keep #1155's fenced layout and add rules for the remaining cases.** Each remaining case (case folding, Unicode normalisation, component length) needs another platform-specific rule and none can be exercised in Linux CI. The layout is the cause, so fencing it further does not converge.
- **Temp files alone, still hashing the on-disk tree.** Not viable: `DirSha256` hashes every entry's basename, so renaming files changes the fingerprint.
- **A CRC64- or ETag-derived fingerprint to avoid downloading.** S3 stores a full-object CRC64NVME for every object uploaded since December 2024, which makes it tempting. Rejected: CRC64 is linear and trivially forgeable, so an attacker who can write the bucket can replace approved content while keeping the fingerprint, which defeats the purpose of a compliance fingerprint. Hashing the CRC with SHA256 adds no resistance. ETag is MD5 of parts for multipart uploads and depends on client chunk size.
- **Incremental snapshots keyed on stored checksums.** Only `VersionId` is a trustworthy skip signal and only with versioning enabled; the design needs persisted state the stateless CLI does not have. Left for a separate decision.

## Consequences

- The fingerprint is the same on every operating system, with Linux semantics. A snapshot run on Windows or macOS against a bucket with case-colliding or backslash keys moves to the Linux value. The same holds for a backslash inside an ignore rule: the virtual glob reads it as an escape, as `filepath.Glob` does on Linux, where on Windows `filepath.Glob` reads it as a separator, so such a rule now resolves differently for an S3 snapshot than for `attest artifact --artifact-type dir` run on the same Windows machine. This is a correction and is release-noted.
- `localPathForS3Key`, `filepath.IsLocal`, the `O_EXCL` open, the `ENOTDIR` branch, `containsSingleFile` and the platform-conditional tests from #1155 are deleted. No code in the S3 path branches on the operating system any more. The codebase does not get smaller, though: the virtual tree, the key rule with its collision reporting, and above all the faithful simulation of `filepathx.Glob` add several hundred lines, most of them owed to reproducing `.kosli_ignore` semantics exactly. That is the price of the compatibility contract, paid once and shared with #1069.
- The bucket's ignore file is recognised by the exact key `.kosli_ignore`. On disk, `ignoreFilePathInTree` also accepted a case-folded spelling such as `.KOSLI_IGNORE` where the operator's filesystem folded case, so on macOS or Windows such a bucket had its rules applied; now it does not, its exclusions stop applying and its fingerprint moves to the Linux value. Same correction as the key case above, and release-noted with it.
- `...`, `.. `, `CON`, colon and backslash keys snapshot again. `..` segments and colliding keys remain errors and now name every key involved.
- Excluded and ignored objects are not downloaded. Each object's temp file is removed once hashed, so peak temp disk falls from the whole bucket to one object, and to the configured budget now that #1167 has landed. The other half of that trade is memory: the old loop streamed the listing a page at a time, while a tree cannot be fingerprinted without holding it, so peak memory rises from one listing page to every object that survives the filters, at a few hundred bytes per object on top of its key across the listing, the key-to-path maps, the manifest and the tree. A million objects is a few hundred megabytes. #1167 sized its budget against this figure.
- #1069 rebases onto the shared layer: metadata mode becomes a sha256 source plugged into the same list, normalise, exclude, tree pipeline, its key rule disappears in favour of rule 2, and its rejection of buckets with a root `.kosli_ignore` becomes a download of that one object.
- Delivery was two pull requests. The first, #1180, carried this decision with sequential downloads, so the security-relevant review was not mixed with performance work. The second, #1167, added parallel downloads, the byte budget and the flags that tune them.
