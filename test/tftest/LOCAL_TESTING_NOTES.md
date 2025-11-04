# Terraform HCL Tests - Local Testing Notes

## Issue with `terraform test -chdir`

**Problem:** The `terraform test` command does NOT support the `-chdir` flag in any Terraform version.

**Solution:** Change to the test directory first, then run `terraform test`:

```bash
# ❌ WRONG - This fails
terraform test -chdir=test/tftest/datastores_datasource

# ✅ CORRECT - Use subshell
(cd test/tftest/datastores_datasource && terraform test)

# ✅ CORRECT - Or change directory
cd test/tftest/datastores_datasource
terraform test
```

## Local Testing Requirements

1. **Terraform 1.6.0+** required for `terraform test` command
   - Check version: `terraform version`
   - If < 1.6.0, tests will say "No tests defined"

2. **Set environment variables:**
   ```bash
   export TF_VAR_pbs_endpoint="https://192.168.1.108:8007"
   export TF_VAR_pbs_username="root@pam"
   export TF_VAR_pbs_password="your-password"
   ```

3. **Build and install provider:**
   ```bash
   go build .
   make install  # Or use run_hcl_tests.sh which does this
   ```

## Using the Helper Script

The `run_hcl_tests.sh` script handles everything:
- Builds provider
- Installs to plugin directory
- Runs tests with proper directory changes
- Works with Terraform 1.6.0+

```bash
./test/tftest/run_hcl_tests.sh                    # All tests
./test/tftest/run_hcl_tests.sh datastores_datasource  # Specific test
```

## CI Testing

CI uses Terraform 1.7.0 and the corrected syntax:
```yaml
(cd test/tftest/datastores_datasource && terraform test)
```

This was fixed in commit [COMMIT_HASH] after discovering `-chdir` is not supported.
