# CI/CD Pipeline for Ansible

Automated testing and deployment workflows using modern Ansible tooling.

## Pipeline Stages

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│    Lint      │───▶│    Test      │───▶│   Deploy     │───▶│   Verify     │
│ ansible-lint │    │   Molecule   │    │   Playbook   │    │   Smoke      │
│  yamllint    │    │   Container  │    │   Rolling    │    │   Tests      │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
```

## ansible-lint

### Installation

```bash
# Via pip
pip install ansible-lint

# Via pipx (isolated)
pipx install ansible-lint

# Specific version
pip install 'ansible-lint>=24.7.0,<25.0.0'
```

### Configuration

```yaml
# .ansible-lint
---
profile: production

# Exclude paths
exclude_paths:
  - .cache/
  - .github/
  - molecule/
  - collections/

# Skip specific rules
skip_list:
  - yaml[line-length]  # Allow long lines
  - name[casing]        # Allow flexible naming

# Warn instead of error
warn_list:
  - experimental
  - role-name[path]

# Enable optional rules
enable_list:
  - no-log-password
  - no-same-owner

# Offline mode (skip galaxy installs)
offline: false

# Use loop instead of with_*
loop_var_prefix: "^(__|{role}_)"
```

### Running

```bash
# Lint everything
ansible-lint

# Lint specific files
ansible-lint playbooks/*.yml roles/*/

# Fix auto-fixable issues
ansible-lint --fix

# Show specific rule info
ansible-lint -L  # List all rules
ansible-lint -T  # List all tags

# Output formats
ansible-lint -f json
ansible-lint -f codeclimate > lint-report.json
```

### Common Rule Fixes

| Rule | Issue | Fix |
|------|-------|-----|
| `name[missing]` | Task without name | Add descriptive name |
| `yaml[truthy]` | `yes/no` instead of bool | Use `true/false` |
| `fqcn[action-core]` | Short module names | Use FQCN: `ansible.builtin.apt` |
| `no-changed-when` | Command without changed_when | Add `changed_when` |
| `risky-shell-pipe` | Unhandled pipe failure | Add `set -o pipefail` |

## yamllint

### Configuration

```yaml
# .yamllint.yml
---
extends: default

rules:
  line-length:
    max: 120
    level: warning
  truthy:
    allowed-values: ['true', 'false', 'yes', 'no']
  comments:
    require-starting-space: true
    min-spaces-from-content: 1
  braces:
    min-spaces-inside: 0
    max-spaces-inside: 1
  brackets:
    min-spaces-inside: 0
    max-spaces-inside: 0
  indentation:
    spaces: 2
    indent-sequences: true
```

### Running

```bash
yamllint .
yamllint -f colored .
yamllint -d "{extends: relaxed, rules: {line-length: {max: 120}}}" .
```

## Molecule

### Installation

```bash
pip install 'molecule[docker]'

# With specific drivers
pip install 'molecule[docker,podman,vagrant]'
```

### Project Structure

```
roles/baseline/
├── tasks/
├── handlers/
├── defaults/
├── meta/
└── molecule/
    ├── default/              # Default scenario
    │   ├── molecule.yml      # Configuration
    │   ├── converge.yml      # Apply role
    │   ├── verify.yml        # Test assertions
    │   ├── prepare.yml       # Pre-test setup (optional)
    │   └── cleanup.yml       # Post-test cleanup (optional)
    └── rocky/                # Alternative scenario
        ├── molecule.yml
        └── converge.yml
```

### Configuration

```yaml
# molecule/default/molecule.yml
---
dependency:
  name: galaxy
  options:
    requirements-file: requirements.yml

driver:
  name: docker

platforms:
  - name: ubuntu-24
    image: ubuntu:24.04
    pre_build_image: true
    command: /lib/systemd/systemd
    privileged: true
    volumes:
      - /sys/fs/cgroup:/sys/fs/cgroup:rw
    cgroupns_mode: host

  - name: ubuntu-22
    image: ubuntu:22.04
    pre_build_image: true
    command: /lib/systemd/systemd
    privileged: true
    volumes:
      - /sys/fs/cgroup:/sys/fs/cgroup:rw
    cgroupns_mode: host

provisioner:
  name: ansible
  playbooks:
    prepare: prepare.yml
    converge: converge.yml
    verify: verify.yml
  inventory:
    host_vars:
      ubuntu-24:
        baseline__ssh_port: 22
      ubuntu-22:
        baseline__ssh_port: 22
  config_options:
    defaults:
      callbacks_enabled: profile_tasks
      stdout_callback: yaml

verifier:
  name: ansible

scenario:
  name: default
  test_sequence:
    - dependency
    - cleanup
    - destroy
    - syntax
    - create
    - prepare
    - converge
    - idempotence
    - verify
    - cleanup
    - destroy
```

### Converge Playbook

```yaml
# molecule/default/converge.yml
---
- name: Converge
  hosts: all
  become: true

  tasks:
    - name: Include baseline role
      ansible.builtin.include_role:
        name: baseline
```

### Verify Playbook

```yaml
# molecule/default/verify.yml
---
- name: Verify
  hosts: all
  become: true
  gather_facts: true

  tasks:
    - name: Check SSH configuration
      ansible.builtin.command: sshd -T
      register: sshd_config
      changed_when: false

    - name: Assert password auth disabled
      ansible.builtin.assert:
        that:
          - "'passwordauthentication no' in sshd_config.stdout | lower"
        fail_msg: "Password authentication is still enabled"

    - name: Check UFW status
      ansible.builtin.command: ufw status
      register: ufw_status
      changed_when: false

    - name: Assert firewall is active
      ansible.builtin.assert:
        that:
          - "'Status: active' in ufw_status.stdout"
        fail_msg: "UFW is not active"

    - name: Check fail2ban status
      ansible.builtin.command: fail2ban-client status sshd
      register: fail2ban_status
      changed_when: false
      failed_when: false

    - name: Assert fail2ban is running
      ansible.builtin.assert:
        that:
          - fail2ban_status.rc == 0
        fail_msg: "fail2ban is not running"
```

### Running Molecule

```bash
# Full test sequence
molecule test

# Specific scenario
molecule test -s rocky

# Interactive development
molecule create    # Create containers
molecule converge  # Apply role
molecule verify    # Run tests
molecule login     # SSH into container
molecule destroy   # Cleanup

# Skip destroy for debugging
molecule test --destroy=never

# List scenarios
molecule list

# Lint only
molecule lint
```

## Pre-commit Hooks

### Configuration

```yaml
# .pre-commit-config.yaml
---
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
        args: ['--unsafe']  # Allow Jinja2 in YAML
      - id: check-added-large-files
      - id: check-merge-conflict

  - repo: https://github.com/adrienverge/yamllint
    rev: v1.35.1
    hooks:
      - id: yamllint
        args: ['-c', '.yamllint.yml']

  - repo: https://github.com/ansible/ansible-lint
    rev: v24.7.0
    hooks:
      - id: ansible-lint
        args: ['--fix']
        additional_dependencies:
          - ansible-core>=2.15

  - repo: local
    hooks:
      - id: molecule-test
        name: Molecule Test
        entry: molecule test --destroy=always
        language: system
        pass_filenames: false
        stages: [push]  # Only on push, not commit
```

### Setup

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install
pre-commit install --hook-type pre-push

# Run manually
pre-commit run --all-files

# Update hooks
pre-commit autoupdate
```

## GitHub Actions

### Complete Workflow

```yaml
# .github/workflows/ansible.yml
---
name: Ansible CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.12'

      - name: Install dependencies
        run: |
          pip install ansible-lint yamllint

      - name: Run yamllint
        run: yamllint .

      - name: Run ansible-lint
        run: ansible-lint --strict

  molecule:
    runs-on: ubuntu-latest
    needs: lint
    strategy:
      fail-fast: false
      matrix:
        role:
          - baseline
          - app
        scenario:
          - default
          - rocky

    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.12'

      - name: Install dependencies
        run: |
          pip install 'molecule[docker]' ansible-core

      - name: Install Galaxy dependencies
        run: |
          ansible-galaxy install -r requirements.yml
          ansible-galaxy collection install -r requirements.yml

      - name: Run Molecule
        run: molecule test -s ${{ matrix.scenario }}
        working-directory: roles/${{ matrix.role }}

  deploy:
    runs-on: ubuntu-latest
    needs: molecule
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment: production

    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.12'

      - name: Install Ansible
        run: pip install ansible-core

      - name: Install dependencies
        run: |
          ansible-galaxy install -r requirements.yml
          ansible-galaxy collection install -r requirements.yml

      - name: Create vault password file
        run: echo "${{ secrets.VAULT_PASSWORD }}" > .vault_pass

      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.SSH_PRIVATE_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          ssh-keyscan -H ${{ secrets.SSH_HOST }} >> ~/.ssh/known_hosts

      - name: Run playbook (check mode first)
        run: |
          ansible-playbook playbooks/site.yml \
            -i inventories/prod \
            --vault-password-file=.vault_pass \
            --check --diff

      - name: Run playbook
        run: |
          ansible-playbook playbooks/site.yml \
            -i inventories/prod \
            --vault-password-file=.vault_pass

      - name: Cleanup
        if: always()
        run: |
          rm -f .vault_pass
          rm -f ~/.ssh/id_rsa
```

## GitLab CI

```yaml
# .gitlab-ci.yml
---
stages:
  - lint
  - test
  - deploy

variables:
  PIP_CACHE_DIR: "$CI_PROJECT_DIR/.pip-cache"
  ANSIBLE_FORCE_COLOR: "true"

.ansible-base:
  image: python:3.12
  before_script:
    - pip install ansible-core 'molecule[docker]' ansible-lint yamllint
    - ansible-galaxy install -r requirements.yml
    - ansible-galaxy collection install -r requirements.yml

lint:
  stage: lint
  extends: .ansible-base
  script:
    - yamllint .
    - ansible-lint --strict
  cache:
    paths:
      - .pip-cache/

molecule:
  stage: test
  extends: .ansible-base
  services:
    - docker:dind
  variables:
    DOCKER_HOST: tcp://docker:2375
  script:
    - cd roles/baseline && molecule test
  parallel:
    matrix:
      - MOLECULE_DISTRO: [ubuntu2404, ubuntu2204, rocky9]

deploy:
  stage: deploy
  extends: .ansible-base
  only:
    - main
  environment:
    name: production
  script:
    - echo "$VAULT_PASSWORD" > .vault_pass
    - echo "$SSH_PRIVATE_KEY" > ~/.ssh/id_rsa
    - chmod 600 ~/.ssh/id_rsa
    - ansible-playbook playbooks/site.yml -i inventories/prod --vault-password-file=.vault_pass
  after_script:
    - rm -f .vault_pass ~/.ssh/id_rsa
```

## Execution Environments

For fully reproducible execution:

```yaml
# execution-environment.yml
---
version: 3

build_arg_defaults:
  ANSIBLE_GALAXY_CLI_COLLECTION_OPTS: '--pre'

dependencies:
  galaxy: requirements.yml
  python:
    - boto3>=1.28.0
    - botocore>=1.31.0
    - jmespath>=1.0.0
  system:
    - openssh-clients
    - sshpass

images:
  base_image:
    name: quay.io/ansible/ansible-runner:latest

additional_build_steps:
  prepend_final:
    - RUN pip install --upgrade pip
  append_final:
    - RUN ansible --version
```

Build and use:

```bash
# Build
ansible-builder build -t my-ee:latest

# Push to registry
docker push registry.example.com/my-ee:latest

# Use with ansible-navigator
ansible-navigator run playbooks/site.yml \
  -i inventories/prod \
  --execution-environment-image my-ee:latest \
  --mode stdout
```

## Best Practices Summary

1. **Lint on every commit** - Pre-commit hooks
2. **Test roles with Molecule** - Multiple platforms
3. **Check mode before apply** - `--check --diff`
4. **Separate environments** - Different vault passwords
5. **Version control everything** - Including CI configs
6. **Use execution environments** - Reproducible builds
7. **Automate deployments** - GitOps workflow
8. **Monitor pipelines** - Alert on failures
