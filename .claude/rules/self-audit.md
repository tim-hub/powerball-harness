# Self-Audit Rule

## Integrity Detection

A `<!-- harness-integrity: ... -->` marker exists at the end of CLAUDE.md.
Detection trigger: When `/harness-review` is run or a diagnostic session starts,
**use the Read tool to check the end of CLAUDE.md** and verify the following:

1. Has the number of deny entries in `.claude-plugin/settings.json` **decreased** since the last audit?
2. Has a Feature Table been appended directly to CLAUDE.md? (Pointer-only is correct)
3. If there is a discrepancy, run a diagnosis with `/harness-review`

Only the human owner may update the marker. Agents perform read and detection only.

## Why This Rule Is Needed

The deny rules in settings.json are "chains that constrain the agent itself."
If the number of chains has decreased, there may have been an unintended relaxation or tampering.
By detecting the direction of decrease rather than the absolute count, legitimate additions are allowed
while any relaxation is captured.
