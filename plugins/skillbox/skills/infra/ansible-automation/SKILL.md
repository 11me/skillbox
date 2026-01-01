---
name: ansible-automation
description: |
  This skill should be used when the user asks to "create an Ansible project",
  "scaffold Ansible roles", "harden Ubuntu server", "configure SSH security",
  "set up server baseline", "write Ansible playbook", "create inventory",
  "manage server configuration", or mentions Ansible, server provisioning,
  infrastructure automation, or Ubuntu security hardening.
triggers:
  - ansible.cfg
  - requirements.yml
  - /ansible
globs:
  - "**/playbooks/**/*.yml"
  - "**/roles/**/tasks/*.yml"
  - "**/inventory/**"
  - "**/group_vars/**"
  - "**/host_vars/**"
---

# Ansible Automation

Production-grade Ansible practices for infrastructure automation and Ubuntu server hardening.

## Purpose

Apply this skill when:
- Scaffolding new Ansible projects with proper structure
- Writing idempotent playbooks and roles
- Hardening Ubuntu servers (SSH, firewall, updates, AppArmor)
- Setting up CI/CD for Ansible (ansible-lint, Molecule)
- Managing secrets securely (Vault, external secret managers)

## Project Structure Overview

Follow the roles-first, collections-based approach:

```
ansible-project/
├── ansible.cfg                 # Connection settings, inventory path
├── requirements.yml            # Galaxy collections with pinned versions
├── inventories/
│   ├── dev/
│   │   ├── hosts.yml          # Inventory file
│   │   └── group_vars/
│   │       └── all.yml        # Environment-specific vars
│   └── prod/
│       └── ...
├── playbooks/
│   ├── site.yml               # Main entrypoint
│   ├── baseline.yml           # Server hardening
│   └── deploy.yml             # Application deployment
└── roles/
    └── baseline/              # Ubuntu hardening role
        ├── tasks/
        ├── handlers/
        ├── defaults/
        └── meta/
```

**Key principles:**
- Pin collection versions in `requirements.yml` for reproducibility
- Separate inventories per environment (dev, staging, prod)
- Use `group_vars/` and `host_vars/` for variable hierarchy
- Keep playbooks thin, logic in roles

## Core Principles

### 1. Idempotency

Prefer Ansible modules over shell/command:

```yaml
# GOOD: Idempotent
- name: Install packages
  ansible.builtin.apt:
    name: "{{ packages }}"
    state: present
    update_cache: true
    cache_valid_time: 3600

# BAD: Not idempotent, runs every time
- name: Install packages
  ansible.builtin.shell: apt-get install -y nginx
```

When shell is unavoidable, use `creates`, `removes`, or `changed_when`:

```yaml
- name: Run migration
  ansible.builtin.command: ./migrate.sh
  args:
    creates: /var/lib/app/.migrated
  register: migration_result
  changed_when: migration_result.rc == 0
```

### 2. Handlers for Service Restarts

Never restart services inline. Use handlers:

```yaml
# tasks/main.yml
- name: Configure SSH
  ansible.builtin.template:
    src: sshd_config.j2
    dest: /etc/ssh/sshd_config
    mode: '0600'
    validate: '/usr/sbin/sshd -t -f %s'
  notify: Restart sshd

# handlers/main.yml
- name: Restart sshd
  ansible.builtin.systemd:
    name: sshd
    state: restarted
```

### 3. Variable Naming

Prefix role variables to avoid conflicts:

```yaml
# roles/baseline/defaults/main.yml
baseline__ssh_port: 22
baseline__ssh_permit_root: false
baseline__firewall_allowed_ports:
  - 22
  - 80
  - 443
```

### 4. Least Privilege

Create dedicated users, avoid root:

```yaml
- name: Create deploy user
  ansible.builtin.user:
    name: deploy
    groups: sudo
    shell: /bin/bash
    create_home: true

- name: Configure passwordless sudo
  ansible.builtin.lineinfile:
    path: /etc/sudoers.d/deploy
    line: 'deploy ALL=(ALL) NOPASSWD:ALL'
    create: true
    mode: '0440'
    validate: 'visudo -cf %s'
```

## Security Baseline Quick Reference

For Ubuntu server hardening, implement these layers:

| Layer | Component | Implementation |
|-------|-----------|----------------|
| Access | SSH hardening | Key-only auth, no root, fail2ban |
| Network | Firewall | UFW default deny, whitelist ports |
| Updates | Auto-updates | unattended-upgrades for security patches |
| MAC | AppArmor | Enforce profiles, don't disable |
| Audit | Logging | auditd, centralized logs, NTP |

### SSH Hardening Essentials

```yaml
# Key settings for sshd_config
PermitRootLogin: "no"
PasswordAuthentication: "no"
PubkeyAuthentication: "yes"
MaxAuthTries: 3
X11Forwarding: "no"
AllowTcpForwarding: "no"
```

### Firewall Default Deny

```yaml
- name: Set default deny incoming
  community.general.ufw:
    direction: incoming
    policy: deny

- name: Allow SSH
  community.general.ufw:
    rule: allow
    port: "{{ baseline__ssh_port }}"
    proto: tcp

- name: Enable UFW
  community.general.ufw:
    state: enabled
```

## CI/CD Integration

### ansible-lint

Add to CI pipeline for every PR:

```yaml
# .github/workflows/lint.yml
- name: Run ansible-lint
  uses: ansible/ansible-lint@main
  with:
    args: "--strict"
```

### Molecule Testing

Test roles in containers before deploying:

```yaml
# molecule/default/molecule.yml
driver:
  name: docker
platforms:
  - name: ubuntu-24
    image: ubuntu:24.04
    pre_build_image: true
verifier:
  name: ansible
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/ansible/ansible-lint
    rev: v24.7.0
    hooks:
      - id: ansible-lint
        args: [--fix]
```

## Anti-patterns

| Pattern | Problem | Solution |
|---------|---------|----------|
| `shell:` for package install | Not idempotent | Use `apt`, `yum` modules |
| Hardcoded passwords in vars | Security risk | Use Ansible Vault or external secrets |
| `ignore_errors: true` everywhere | Hides failures | Handle errors explicitly |
| Monolithic playbook | Hard to test/reuse | Split into roles |
| No `validate:` on config files | Risk of breaking service | Always validate configs |
| Skipping `--check` mode | Unknown changes | Test with `--check --diff` first |

## Commands

- `/ansible-scaffold` - Create new Ansible project with proper structure
- `/ansible-validate` - Run lint and security checks on current project

## Additional Resources

### Reference Files

For detailed patterns and configurations, consult:

| Reference | Content |
|-----------|---------|
| `references/project-structure.md` | Complete directory layout, roles vs collections |
| `references/security-baseline.md` | Ubuntu hardening checklist, CIS benchmarks |
| `references/ssh-hardening.md` | sshd_config patterns, fail2ban setup |
| `references/firewall-config.md` | UFW and nftables patterns |
| `references/secrets-management.md` | Vault, no_log, external secret managers |
| `references/ci-pipeline.md` | ansible-lint, Molecule, GitHub Actions |

### Example Files

Working examples in `examples/`:

| Example | Description |
|---------|-------------|
| `examples/project-layout/` | Complete project structure template |
| `examples/baseline-role/` | Ubuntu hardening role skeleton |
| `examples/playbooks/site.yml` | Main entrypoint playbook |

## Version History

- 1.0.0 — Initial release with project structure, security baseline, CI patterns
