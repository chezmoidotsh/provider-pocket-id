# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Testing

- `make`: Run default build target
- `make build`: Build the provider binary
- `make reviewable`: Run code generation, linters, and tests (use before committing)
- `make test-integration`: Run integration tests using kind
- `make go.build`: Build the Go binary for local testing

### Local Development

- `make run`: Run provider locally out-of-cluster with debug logging
- `make dev`: Create kind cluster, install CRDs, and start provider controllers in debug mode
- `make dev-clean`: Clean up the development kind cluster

### Code Generation

- `make submodules`: Initialize/update build submodules (run this first on new clone)
- Code generation is handled automatically by `make reviewable`

## Project Architecture

This is a Crossplane Provider for [Pocket ID](https://pocket-id.org/) that manages Users, Groups, and OIDC Clients. Other Pocket ID configuration is intentionally not included as it should be managed as-code.

### Core Structure

- **`apis/v1alpha1/`**: Kubernetes API types and CRDs
  - `providerconfig_types.go`: Provider configuration and authentication
  - `user_types.go`: User management (regular users)
  - `adminuser_types.go`: Admin user management
  - `client_types.go`: OIDC client configuration
  - `*groupbinding_types.go`: User-group and client-group associations
  - All types follow Crossplane managed resource patterns with Spec/Status

### Controllers

- **`internal/controller/`**: Reconciliation logic for each resource type
  - `pocketid.go`: Main controller setup that registers all resource controllers
  - `user/`: Regular user management
  - `adminuser/`: Admin user management
  - `oidcclient/`: OIDC client configuration
  - `*groupbinding/`: User-group and client-group binding controllers
  - Controllers implement Crossplane's managed resource lifecycle (Observe, Create, Update, Delete)

### Provider Pattern

- Uses Crossplane runtime for common provider functionality
- Authentication via ProviderConfig with credential sources (Secret, Environment, etc.)
- Each managed resource follows external-name pattern for mapping to Pocket ID resources
- Scope limited to Users, Groups, and OIDC Clients only - other Pocket ID configuration should be managed as-code

### Key Files

- `cmd/provider/main.go`: Provider entry point
- `package/crossplane.yaml`: Provider package definition
- `package/crds/`: Generated CRDs for installation

## Testing Notes

- Integration tests use kind and are defined in `cluster/local/integration_tests.sh`
- Unit tests follow Go conventions with `*_test.go` files
- Always run `make reviewable` before committing to ensure code passes all checks

## Provider Development Workflow

1. Run `make submodules` after cloning
2. Use `make provider.addtype` to add new resource types
3. Implement controller logic for new types
4. Register controllers in `internal/controller/pocketid.go`
5. Test with `make dev` for local development
6. Run `make reviewable` before committing
