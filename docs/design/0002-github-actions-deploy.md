# 0002 GitHub Actions Deploy

## Context

Deployment should use GitHub Actions and `aws-actions/aws-lambda-deploy` instead of CloudFormation.

## Decisions

1. Use GitHub OIDC for AWS authentication.
    - The workflow uses `aws-actions/configure-aws-credentials` with `id-token: write`.
    - This avoids long-lived AWS access keys in GitHub secrets.

2. Build and deploy a ZIP-based custom runtime.
    - The workflow builds `./cmd/lambda` into `.dist/lambda/bootstrap` for `provided.al2023` on `arm64`.
    - `aws-actions/aws-lambda-deploy` deploys that build directory directly.

3. Keep the workflow focused on function deployment only.
    - `aws-actions/aws-lambda-deploy` updates the Lambda function code and configuration.
    - Function URL creation, CORS, and public permissions are managed separately from this repository.
    - This keeps the workflow smaller and avoids embedding extra AWS resource lifecycle logic in shell steps.

4. Remove CloudFormation from the deployment path.
    - The previous template is deleted to avoid two conflicting deployment sources.
    - The repository now has one deployment contract: GitHub Actions updates the function in place.

5. Do not emit deployment artifacts in workflow output.
    - The workflow intentionally avoids printing the Function URL or deployed
      function ARN into the job summary.
    - This keeps the deployment logs lower-noise and avoids exposing details
      that are not needed for day-to-day personal use.

## Required GitHub configuration

Configure these repository or environment variables before running the workflow:

- `AWS_REGION`
- `AWS_DEPLOY_ROLE_ARN`
- `LAMBDA_FUNCTION_NAME`
- `LAMBDA_EXECUTION_ROLE_ARN`

Recommended target environment name:

- `production`

## IAM notes

The GitHub OIDC role needs enough permissions for:

- Lambda create/update code and configuration
- `iam:PassRole` for the Lambda execution role
