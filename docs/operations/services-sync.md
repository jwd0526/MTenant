# Service Sync Documentation

**Last Updated:** 2026-05-19\
*Initial doc*

This repository uses **git subtree** to sync standalone service repositories with the monorepo.

## Standalone Repositories

- **deal-service**: https://github.com/jwd0526/deal-service
- **tenant-service**: https://github.com/jwd0526/tenant-service

## Syncing Workflow

### Quick Sync (Using Script)

Use the `sync-services.sh` script for easy syncing:

```bash
# Pull changes from standalone repo to monorepo
./sync-services.sh deal pull
./sync-services.sh tenant pull

# Push changes from monorepo to standalone repo
./sync-services.sh deal push
./sync-services.sh tenant push
```

### Manual Git Subtree Commands

If you prefer manual control:

#### Pull changes FROM standalone repo TO monorepo

```bash
# Pull deal-service changes
git subtree pull --prefix=services/deal-service deal-service main --squash

# Pull tenant-service changes
git subtree pull --prefix=services/tenant-service tenant-service main --squash
```

#### Push changes FROM monorepo TO standalone repo

```bash
# Push deal-service changes
git subtree push --prefix=services/deal-service deal-service main

# Push tenant-service changes
git subtree push --prefix=services/tenant-service tenant-service main
```

## Workflow Examples

### Scenario 1: You made changes in the standalone deal-service repo

1. Pull changes to MTenant:
   ```bash
   ./sync-services.sh deal pull
   ```
2. Commit the merge in MTenant if needed
3. Push MTenant to origin

### Scenario 2: You made changes in MTenant's services/deal-service

1. Commit changes in MTenant
2. Push to standalone repo:
   ```bash
   ./sync-services.sh deal push
   ```

### Scenario 3: Working on both simultaneously

It's recommended to make changes in one location at a time to avoid conflicts. If you must work in both:

1. Commit changes in standalone repo
2. Pull to monorepo: `./sync-services.sh deal pull`
3. Make additional changes in monorepo if needed
4. Push back: `./sync-services.sh deal push`

## Important Notes

- The `--squash` flag on pull combines all commits into one, keeping history clean
- Always commit changes before syncing
- Git subtree preserves full file history
- Remotes are already configured:
  - `deal-service` → https://github.com/jwd0526/deal-service.git
  - `tenant-service` → https://github.com/jwd0526/tenant-service.git
