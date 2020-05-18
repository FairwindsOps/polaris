# Polaris Exit Codes for Audit Runs
<dl>
  <dt>Exit 0</dt>
  <dd>Successful exit code</dd>
  <dt>Exit 1</dt>
  <dd>Could not run audit, or application had a failure while running.</dd>
  <dt>Exit 2</dt>
  <dd>Unused</dd>
  <dt>Exit 3</dt>
  <dd>Exiting due to `--set-exit-code-on-danger` being set and at least one danger was found after an audit.</dd>
  <dt>Edit 4</dd>
  <dd>Exiting due to `--set-exit-code-below-score` being set and the audit resulted in a score less than the minimum score value.</dd>
</dl>
