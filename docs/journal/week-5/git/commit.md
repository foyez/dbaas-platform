# Git Conventional Commits

| Tag        | Use for                                    | Example                               |
| ---------- | ------------------------------------------ | ------------------------------------- |
| `feat`     | New feature                                | `feat: add user search`               |
| `fix`      | Bug fix                                    | `fix: prevent null pointer error`     |
| `docs`     | Documentation                              | `docs: update API usage`              |
| `style`    | Formatting, whitespace, no logic change    | `style: format auth module`           |
| `refactor` | Code restructuring without behavior change | `refactor: simplify validation`       |
| `perf`     | Performance improvement                    | `perf: optimize database query`       |
| `test`     | Tests                                      | `test: add login integration tests`   |
| `build`    | Build system/dependencies                  | `build: upgrade webpack`              |
| `infra`    | Cloud/infrastructure                       | `infra: configure Kubernetes ingress` |
| `ci`       | CI/CD configuration                        | `ci: add GitHub Actions workflow`     |
| `chore`    | Maintenance tasks                          | `chore: update dependencies`          |
| `revert`   | Revert a previous commit                   | `revert: revert login change`         |

### Typical format

```text
<type>[optional scope]: <description>

feat(auth): add OAuth login
fix(api): handle missing user ID
chore: update dependencies
docs: add deployment guide
refactor(db): simplify connection handling
test(auth): add token expiration tests
```

You can also add **breaking changes**:

```text
feat(api)!: remove deprecated endpoints
```

The most commonly needed set for a project is:

`feat` · `fix` · `docs` · `refactor` · `test` · `chore` · `ci` · `perf` · `build` · `revert`
