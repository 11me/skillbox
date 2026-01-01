# Ansible Project Structure

Complete guide to organizing Ansible projects for production use.

## Directory Layout

```
ansible-project/
├── ansible.cfg                     # Ansible configuration
├── requirements.yml                # Galaxy dependencies
├── .ansible-lint                   # Linter configuration
├── .pre-commit-config.yaml         # Pre-commit hooks
├── .yamllint.yml                   # YAML linter config
│
├── inventories/                    # Environment-specific inventories
│   ├── dev/
│   │   ├── hosts.yml              # Inventory definition
│   │   ├── group_vars/
│   │   │   ├── all.yml            # Vars for all hosts
│   │   │   └── webservers.yml     # Vars for webservers group
│   │   └── host_vars/
│   │       └── server1.yml        # Host-specific vars
│   ├── staging/
│   │   └── ...
│   └── prod/
│       └── ...
│
├── playbooks/                      # Entry point playbooks
│   ├── site.yml                   # Main orchestration
│   ├── baseline.yml               # Server hardening
│   ├── deploy.yml                 # Application deployment
│   └── maintenance.yml            # Maintenance tasks
│
├── roles/                          # Local roles
│   ├── baseline/                  # Server hardening
│   │   ├── tasks/
│   │   │   ├── main.yml
│   │   │   ├── users.yml
│   │   │   ├── ssh.yml
│   │   │   ├── firewall.yml
│   │   │   └── updates.yml
│   │   ├── handlers/
│   │   │   └── main.yml
│   │   ├── templates/
│   │   │   └── sshd_config.j2
│   │   ├── defaults/
│   │   │   └── main.yml
│   │   ├── vars/
│   │   │   └── main.yml
│   │   ├── meta/
│   │   │   └── main.yml
│   │   └── molecule/              # Role tests
│   │       └── default/
│   │           ├── molecule.yml
│   │           ├── converge.yml
│   │           └── verify.yml
│   └── app/                       # Application role
│       └── ...
│
├── collections/                    # Local collections (optional)
│   └── requirements.yml           # Can symlink to root
│
├── filter_plugins/                 # Custom Jinja2 filters
├── library/                        # Custom modules
├── lookup_plugins/                 # Custom lookups
│
└── molecule/                       # Project-level tests
    └── default/
        ├── molecule.yml
        └── converge.yml
```

## Key Files

### ansible.cfg

```ini
[defaults]
# Inventory location
inventory = inventories/dev

# Performance
forks = 20
pipelining = true

# SSH settings
host_key_checking = false
remote_user = deploy

# Logging
log_path = logs/ansible.log
callback_whitelist = timer, profile_tasks

# Roles path
roles_path = roles:~/.ansible/roles:/usr/share/ansible/roles

# Collections path
collections_path = collections:~/.ansible/collections

# Retry files
retry_files_enabled = false

# Fact caching (optional)
gathering = smart
fact_caching = jsonfile
fact_caching_connection = /tmp/ansible_facts
fact_caching_timeout = 86400

[privilege_escalation]
become = true
become_method = sudo
become_user = root
become_ask_pass = false

[ssh_connection]
ssh_args = -o ControlMaster=auto -o ControlPersist=60s -o StrictHostKeyChecking=no
control_path = %(directory)s/%%h-%%r
pipelining = true

[inventory]
enable_plugins = yaml, ini, auto
```

### requirements.yml

Pin all collection versions for reproducibility:

```yaml
---
collections:
  # Community collections
  - name: community.general
    version: ">=8.0.0,<9.0.0"
  - name: community.crypto
    version: ">=2.0.0,<3.0.0"

  # Cloud providers
  - name: amazon.aws
    version: ">=7.0.0,<8.0.0"
  - name: google.cloud
    version: ">=1.3.0,<2.0.0"

  # Container/Kubernetes
  - name: kubernetes.core
    version: ">=3.0.0,<4.0.0"

  # Security
  - name: ansible.posix
    version: ">=1.5.0,<2.0.0"

  # Specific version pin
  - name: community.docker
    version: "3.8.0"

roles:
  # Galaxy roles (prefer collections)
  - name: geerlingguy.docker
    version: "6.2.0"
```

Install dependencies:

```bash
ansible-galaxy install -r requirements.yml
ansible-galaxy collection install -r requirements.yml
```

## Inventory Patterns

### YAML Inventory (Recommended)

```yaml
# inventories/prod/hosts.yml
---
all:
  children:
    webservers:
      hosts:
        web1.example.com:
        web2.example.com:
      vars:
        http_port: 80

    databases:
      hosts:
        db1.example.com:
          postgres_version: 16
        db2.example.com:
          postgres_version: 16

    loadbalancers:
      hosts:
        lb1.example.com:

  vars:
    ansible_user: deploy
    ansible_python_interpreter: /usr/bin/python3
```

### Dynamic Inventory

For cloud environments, use dynamic inventory plugins:

```yaml
# inventories/aws/aws_ec2.yml
---
plugin: amazon.aws.aws_ec2
regions:
  - us-east-1
  - eu-west-1
filters:
  tag:Environment: production
keyed_groups:
  - key: tags.Role
    prefix: role
  - key: placement.availability_zone
    prefix: az
hostnames:
  - private-ip-address
compose:
  ansible_host: private_ip_address
```

## Variable Hierarchy

Variables are loaded in order of precedence (lowest to highest):

1. `role/defaults/main.yml` - Role defaults
2. `inventory/group_vars/all.yml` - All hosts
3. `inventory/group_vars/<group>.yml` - Group-specific
4. `inventory/host_vars/<host>.yml` - Host-specific
5. Play vars and vars_files
6. Role vars (`role/vars/main.yml`)
7. Block vars, task vars
8. Extra vars (`-e "key=value"`)

### group_vars Example

```yaml
# inventories/prod/group_vars/all.yml
---
# Environment
env: prod
domain: example.com

# SSH
ssh_port: 22
ssh_allow_groups:
  - ssh-users
  - deploy

# Updates
auto_updates_enabled: true
auto_updates_reboot: false

# Monitoring
monitoring_enabled: true
prometheus_endpoint: http://prometheus.internal:9090
```

### host_vars Example

```yaml
# inventories/prod/host_vars/db1.yml
---
# Override for specific host
postgres_max_connections: 200
postgres_shared_buffers: "4GB"

# Host-specific secrets (prefer Vault)
postgres_password: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  ...
```

## Roles vs Collections

### When to Use Roles

- Internal, project-specific logic
- Simple, single-purpose automation
- Tight coupling to project structure
- Development/testing in progress

### When to Use Collections

- Reusable across multiple projects
- Sharing with community/organization
- Multiple related roles, plugins, modules
- Versioned releases

### Collection Structure

```
my_namespace/my_collection/
├── galaxy.yml               # Collection metadata
├── plugins/
│   ├── modules/
│   ├── module_utils/
│   ├── filter/
│   └── lookup/
├── roles/
│   └── my_role/
├── playbooks/
├── docs/
└── tests/
```

## Tags Strategy

Use tags for selective execution:

```yaml
# playbooks/site.yml
---
- name: Configure servers
  hosts: all

  tasks:
    - name: Include baseline role
      ansible.builtin.include_role:
        name: baseline
      tags:
        - baseline
        - security

    - name: Include application role
      ansible.builtin.include_role:
        name: app
      tags:
        - app
        - deploy
```

Run specific tags:

```bash
# Only hardening
ansible-playbook playbooks/site.yml --tags security

# Skip deployment
ansible-playbook playbooks/site.yml --skip-tags deploy

# List available tags
ansible-playbook playbooks/site.yml --list-tags
```

## Environment Separation

### Method 1: Separate Inventories (Recommended)

```bash
# Development
ansible-playbook -i inventories/dev playbooks/site.yml

# Production
ansible-playbook -i inventories/prod playbooks/site.yml
```

### Method 2: Variable Override

```bash
ansible-playbook playbooks/site.yml -e "env=prod"
```

### Method 3: Vault per Environment

```bash
# Dev vault
ansible-playbook playbooks/site.yml \
  --vault-password-file=.vault_pass_dev

# Prod vault
ansible-playbook playbooks/site.yml \
  --vault-password-file=.vault_pass_prod
```

## Execution Environments

For reproducible execution across CI and local development:

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
  system:
    - openssh-clients
    - sshpass
images:
  base_image:
    name: quay.io/ansible/ansible-runner:latest
```

Build and use:

```bash
# Build execution environment
ansible-builder build -t my-ee:latest

# Run with ansible-navigator
ansible-navigator run playbooks/site.yml -i inventories/prod \
  --execution-environment-image my-ee:latest
```

## Best Practices Summary

1. **Version control everything** - Git repo with code review
2. **Pin dependencies** - Lock collection/role versions
3. **Separate environments** - Different inventories, vaults
4. **Use roles** - Keep playbooks thin
5. **Prefix variables** - Avoid naming conflicts
6. **Test first** - Molecule for roles, `--check --diff` for playbooks
7. **Encrypt secrets** - Vault or external secret manager
8. **Log executions** - Audit trail for compliance
9. **Use collections** - Modern, FQCN syntax
10. **Execution environments** - Reproducible builds
