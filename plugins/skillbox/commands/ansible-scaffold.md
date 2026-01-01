---
name: ansible-scaffold
description: Create new Ansible project with proper structure and optional baseline role
allowed-tools: Read, Write, Glob, AskUserQuestion, Bash
---

# /ansible-scaffold

Create a new Ansible project with production-ready structure, following modern best practices.

## Workflow

### Step 1: Gather Project Information

Use AskUserQuestion to collect:

1. **Project name** - Name for the Ansible project directory
2. **Environments** - Which environments to create (default: dev, prod)
3. **Include baseline role?** - Add Ubuntu hardening role (yes/no)
4. **CI provider** - GitHub Actions, GitLab CI, or none

### Step 2: Create Directory Structure

Create the following structure:

```
{project-name}/
├── ansible.cfg
├── requirements.yml
├── .ansible-lint
├── .yamllint.yml
├── .gitignore
├── inventories/
│   ├── dev/
│   │   ├── hosts.yml
│   │   └── group_vars/
│   │       └── all.yml
│   └── prod/
│       ├── hosts.yml
│       └── group_vars/
│           └── all.yml
├── playbooks/
│   └── site.yml
└── roles/
    └── (baseline/ if requested)
```

### Step 3: Generate Configuration Files

1. **ansible.cfg** - Copy from `skills/infra/ansible-automation/examples/project-layout/ansible.cfg`
2. **requirements.yml** - Copy from `skills/infra/ansible-automation/examples/project-layout/requirements.yml`

3. **.ansible-lint**:
```yaml
---
profile: production
exclude_paths:
  - .cache/
  - .github/
  - molecule/
skip_list:
  - yaml[line-length]
warn_list:
  - experimental
```

4. **.yamllint.yml**:
```yaml
---
extends: default
rules:
  line-length:
    max: 120
    level: warning
  truthy:
    allowed-values: ['true', 'false', 'yes', 'no']
```

5. **.gitignore**:
```
*.retry
.vault_pass*
*.pyc
__pycache__/
.cache/
logs/
```

### Step 4: Create Inventories

For each environment, create:

1. **hosts.yml** - Basic inventory structure
2. **group_vars/all.yml** - Environment-specific variables

Use `skills/infra/ansible-automation/examples/project-layout/inventories/` as templates.

### Step 5: Create Main Playbook

Create `playbooks/site.yml` from `skills/infra/ansible-automation/examples/playbooks/site.yml`.

### Step 6: Add Baseline Role (if requested)

If user selected baseline role:

1. Copy entire `skills/infra/ansible-automation/examples/baseline-role/` to `roles/baseline/`
2. Update variable prefix if project has specific naming

### Step 7: Add CI Pipeline (if requested)

**GitHub Actions:**
```yaml
# .github/workflows/ansible.yml
---
name: Ansible CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install ansible-lint yamllint
      - run: yamllint .
      - run: ansible-lint --strict
```

**GitLab CI:**
```yaml
# .gitlab-ci.yml
---
stages:
  - lint

lint:
  image: python:3.12
  stage: lint
  script:
    - pip install ansible-lint yamllint
    - yamllint .
    - ansible-lint --strict
```

### Step 8: Initialize Git Repository

```bash
git init
git add .
git commit -m "feat: initialize Ansible project structure"
```

## Output

After completion, display:

1. **Summary of created structure**
2. **Next steps:**
   - Add SSH public keys to `baseline__ssh_public_keys`
   - Configure inventory hosts
   - Run: `ansible-galaxy install -r requirements.yml`
   - Test with: `ansible-playbook playbooks/site.yml -i inventories/dev --check --diff`

## Example Usage

```
/ansible-scaffold
```

Creates interactive scaffold with prompts for configuration.

```
/ansible-scaffold myproject
```

Creates project named "myproject" with default options.
