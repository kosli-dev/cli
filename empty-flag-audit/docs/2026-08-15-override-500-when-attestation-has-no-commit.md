# Bug: `attest override --commit` returns 500 when the attestation it overrides has no commit

`kosli attest override` fails with HTTP 500 whenever `--commit` is given and the
attestation being overridden was reported without one. The override does not
happen, and the command exits non-zero after retrying the 500 three times.

Giving `--commit` is the only trigger. The same override with no `--commit`
succeeds against the same attestation, in a fortieth of the time.

## Reproducing it

Against a freshly reset local server:

```
$ kosli create flow probe-ov2 --use-empty-template
$ kosli begin trail probe-ov2-tr --flow probe-ov2
$ kosli attest generic --name t1 --flow probe-ov2 --trail probe-ov2-tr
generic attestation 't1' is reported to trail: probe-ov2-tr    # no --commit

$ kosli attest override --flow probe-ov2 --trail probe-ov2-tr --name t1 \
    --new-compliance-status=true --original-attestation-type generic \
    --reason probe --commit=HEAD --debug
[debug] POST .../trail/probe-ov2-tr/override (status: 500): retrying in 1s (3 left)
[debug] POST .../trail/probe-ov2-tr/override (status: 500): retrying in 2s (2 left)
[debug] POST .../trail/probe-ov2-tr/override (status: 500): retrying in 4s (1 left)
Error: [kosli attest override flow=probe-ov2 trail=probe-ov2-tr] Post ".../override":
       giving up after 4 attempt(s)
```

The same command without `--commit=HEAD` succeeds:

```
attestation 't1' has been overridden in trail: probe-ov2-tr
```

## What decides it

The attestation being overridden, not the override. Report the original
attestation with a commit and the override with `--commit` then works:

```
$ kosli attest generic --name t1 --flow probe-ov3 --trail probe-ov3-tr --commit HEAD
generic attestation 't1' is reported to trail: probe-ov3-tr

$ kosli attest override ... --commit=HEAD
attestation 't1' has been overridden in trail: probe-ov3-tr
```

So the failing case is an attestation reported without git provenance, later
overridden by someone who does pass a commit. Reporting an attestation without
one is ordinary: `--commit` is optional on the attest commands, and the audit's
own baseline for `attest generic` does not pass it.

## The server side

From the local server's log for the failing request:

```
File "/app/src/fastapi_app/v2/attestation.py", line 516, in post_override_attestation
  return common_attestations.add(payload, None, org, flow_name, trail_name, user)
File "/app/src/fastapi_app/common/attestation.py", line 238, in add
  attestations.override(attestation, time_now(), repo=repo)
File "/app/src/model/attestations.py", line 331, in override
  attestation.relevant_commits.append(attestation.git_commit.sha1)
AttributeError: 'NoneType' object has no attribute 'append'
```

`relevant_commits` is `None` on an attestation that was reported without a
commit, and `override` appends to it without checking. The guard is missing on
the server, so no CLI change fixes this for a customer calling the API.

## Why the audit did not report it

The audit ran this combination and recorded it as `accepted`. That is correct
for the question it asks: it compares the run with `--commit` emptied against
the runs with the flag omitted and set, and an empty `--commit` behaves like
omitting it. The 500 happens only in the run with a real commit, which is the
control rather than the measurement.

What found it was the time limit on a command-and-flag pair. The combination
took 7.6 seconds where the median is 0.3, and the seven seconds are the client
waiting out its own retries: one second, then two, then four. A failure the
audit classified as a working control was visible only as time spent.

That is worth remembering about the retry policy generally. Retrying a 500 that
will never succeed turns an instant error into a seven-second one, and the
places where this CLI is slow are mostly places where something failed.
