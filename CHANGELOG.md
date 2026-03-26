# Changelog

## 0.1.0 (2026-03-26)

### Added

- `luckperms_group` resource with meta attributes (display_name, weight, prefix, suffix)
- `luckperms_group_nodes` resource for permission and inheritance nodes
- `luckperms_track` resource for promotion/demotion paths
- `luckperms_group`, `luckperms_groups`, `luckperms_track`, `luckperms_tracks` data sources
- Full import support for all resources
- Meta-node validation (plan-time error for meta nodes in group_nodes)
- `generate` CLI tool for bootstrapping .tf files from existing state
- Docker Compose setup for local development
- CI/CD with GitHub Actions (test + release)
