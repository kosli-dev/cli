---
title: "20260911 - Fingerprint S3 buckets from a virtual tree; object keys never become local paths"
description: "Download each object to an anonymous temp file, hash it, delete it, and compute the directory fingerprint from (key, sha256) pairs so that no S3 key is ever used as a filename"
status: "Proposed"
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

2. **The key becomes a virtual path by one rule, shared by both modes:** `path.Clean(strings.TrimLeft(key, "/"))`, rejecting a result of `.`, `..` or a leading `../`. This is exactly what `filepath.Join` produced on Linux, so `a//b`, `./c.txt` and `/lead.txt` land where they do today. Everything else, including `...`, `.. `, `CON`, colons and backslashes, is an ordinary name because nothing is created under it. The rule exists for fingerprint stability, not safety.

3. **Collisions are data errors that name every key involved.** Two keys resolving to one path, and a key under a prefix that is also an object, cannot be represented as a tree. They are detected before the digest package sees them and reported together, capped like `combineUnusableObjectErrors`, with the existing advice to use `--exclude-regex` or narrow the include filter.

4. **A root `.kosli_ignore` is honoured virtually.** Its rules are parsed with the same comment and whitespace handling as `excludePathsFromFile` and matched against virtual paths with a `**`-capable segment matcher. A matched directory drops its subtree. The ignore file can never exclude itself, as in `DirSha256`. Excluded and ignored objects are not downloaded at all. Equivalence with `DirSha256` on a materialised tree is asserted for every ignore-file case the digest tests already hold.

5. **Downloads run in parallel** behind a bounded semaphore and a bytes-in-flight budget derived from listing sizes, with results written by index for determinism, a cancellable context so the first transport error stops in-flight multipart downloads, and per-key errors collected rather than aborting.

6. **The switch is pinned, not argued.** Fingerprints of representative fake buckets are recorded against `main` before the implementation changes and asserted afterwards, alongside `TestGetS3DataFromClientKeepsTodaysLayoutForUnusualKeys` from #1155, which must stay green through the change.

A throwaway equivalence test run on 2026-09-11 confirmed that `VirtualDirSha256` fed by rule 2 reproduces today's on-disk fingerprint for the #1155 unusual-keys bucket, the `a.txt` beside `a/z` ordering case, nested prefixes with folder markers, and the single-object case including its basename artifact name.

## Alternatives considered

- **Keep #1155's fenced layout and add rules for the remaining cases.** Each remaining case (case folding, Unicode normalisation, component length) needs another platform-specific rule and none can be exercised in Linux CI. The layout is the cause, so fencing it further does not converge.
- **Temp files alone, still hashing the on-disk tree.** Not viable: `DirSha256` hashes every entry's basename, so renaming files changes the fingerprint.
- **A CRC64- or ETag-derived fingerprint to avoid downloading.** S3 stores a full-object CRC64NVME for every object uploaded since December 2024, which makes it tempting. Rejected: CRC64 is linear and trivially forgeable, so an attacker who can write the bucket can replace approved content while keeping the fingerprint, which defeats the purpose of a compliance fingerprint. Hashing the CRC with SHA256 adds no resistance. ETag is MD5 of parts for multipart uploads and depends on client chunk size.
- **Incremental snapshots keyed on stored checksums.** Only `VersionId` is a trustworthy skip signal and only with versioning enabled; the design needs persisted state the stateless CLI does not have. Left for a separate decision.

## Consequences

- The fingerprint is the same on every operating system, with Linux semantics. A snapshot run on Windows or macOS against a bucket with case-colliding or backslash keys moves to the Linux value. This is a correction and is release-noted.
- `localPathForS3Key`, `filepath.IsLocal`, the `O_EXCL` open, the `ENOTDIR` branch, `containsSingleFile` and the platform-conditional tests from #1155 are deleted. The codebase ends with less code than before #1155.
- `...`, `.. `, `CON`, colon and backslash keys snapshot again. `..` segments and colliding keys remain errors and now name every key involved.
- Excluded and ignored objects are not downloaded, and disk use is bounded by the in-flight budget rather than the bucket size.
- #1069 rebases onto the shared layer: metadata mode becomes a sha256 source plugged into the same list, normalise, exclude, tree pipeline, its key rule disappears in favour of rule 2, and its rejection of buckets with a root `.kosli_ignore` becomes a download of that one object.
- Delivery is sliced in `TODO.md` (local, gitignored): lift `VirtualDirSha256` with its tests; add the key rule and collision errors; add virtual ignore rules; switch content mode to temp files sequentially with pinned fingerprints; parallelise.
