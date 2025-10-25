# Code Coverage Setup with Codecov

This repository uses [Codecov](https://codecov.io) to track and display code coverage from both unit tests and integration tests.

## Setup

### 1. Enable Codecov for Your Repository

1. Visit [codecov.io](https://codecov.io) and sign in with your GitHub account
2. Add your repository to Codecov (it will automatically detect it if you have push access)
3. Codecov will provide you with an upload token (optional for public repos)

### 2. Configure Codecov Token (Optional)

For private repositories or to ensure uploads work reliably:

1. Go to your repository on Codecov and find your upload token
2. In your GitHub repository, go to Settings → Secrets and variables → Actions
3. Add a new repository secret:
   - Name: `CODECOV_TOKEN`
   - Value: Your Codecov upload token

### 3. Add Token to Workflow (Optional)

If you added the `CODECOV_TOKEN` secret, update the workflow files to use it:

```yaml
- name: Report coverage
  uses: codecov/codecov-action@v4
  with:
    token: ${{ secrets.CODECOV_TOKEN }}  # Add this line
    files: ./coverage-unit.out
    flags: unit-tests
    name: unit-tests-coverage
```

## Coverage Reports

### Viewing Coverage on Pull Requests

When you create a pull request, Codecov will automatically:
- Post a comment with coverage statistics
- Add status checks showing coverage changes
- Display line-by-line coverage in the "Files changed" tab
- Show which lines are covered (green), uncovered (red), or partially covered (yellow)

### Viewing Coverage Dashboard

Visit your project on Codecov to see:
- Overall coverage percentage
- Coverage trends over time
- File-by-file coverage breakdown
- Coverage comparison between branches
- Pull request coverage impact

## Coverage Files Generated

The workflows generate several coverage files:

- `coverage-unit.out` - Unit test coverage
- `coverage-quick-datastore.out` - Quick smoke and datastore integration tests
- `coverage-aws.out` - AWS S3 integration tests
- `coverage-b2.out` - Backblaze B2 integration tests
- `coverage-scaleway.out` - Scaleway integration tests
- `coverage-vm-integration.out` - Full VM integration tests

Each is uploaded as a separate artifact and reported to Codecov with appropriate flags.

## Coverage Flags

Codecov uses flags to organize coverage reports:

- `unit-tests` - Unit test coverage
- `integration-quick-datastore` - Quick and datastore tests
- `integration-s3-aws` - AWS S3 tests
- `integration-s3-b2` - Backblaze B2 tests
- `integration-s3-scaleway` - Scaleway tests
- `vm-integration-tests` - Full VM integration tests

You can filter and compare coverage by these flags in the Codecov dashboard.

## Customizing Coverage

To customize coverage behavior, create a `.codecov.yml` file in the repository root:

```yaml
coverage:
  status:
    project:
      default:
        target: 70%        # Minimum coverage target
        threshold: 1%      # Allow coverage to drop by this much
    patch:
      default:
        target: 80%        # Require 80% coverage on changed code

comment:
  layout: "header, diff, files"
  behavior: default
  require_changes: false

flags:
  unit-tests:
    paths:
      - fwprovider/
      - pbs/
  integration-tests:
    paths:
      - test/integration/
```

## Running Coverage Locally

To generate coverage locally:

```bash
# Unit tests
go test -coverprofile=coverage.out -covermode=atomic ./...

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out
```

## Troubleshooting

### Coverage Not Uploading

If coverage isn't appearing in Codecov:

1. Check that the Codecov action is running in your workflow
2. Verify the coverage file paths are correct
3. For private repos, ensure `CODECOV_TOKEN` is set
4. Check Codecov's status page for service issues

### Coverage Appears Low

If coverage seems lower than expected:

1. Make sure all test files are being run
2. Check that `-covermode=atomic` is set (handles concurrent tests)
3. Verify coverage files are being generated (check workflow artifacts)
4. Review which files are being tested with `go tool cover -func=coverage.out`

## Additional Resources

- [Codecov Documentation](https://docs.codecov.com/)
- [Codecov GitHub Action](https://github.com/codecov/codecov-action)
- [Go Coverage Documentation](https://go.dev/blog/cover)
