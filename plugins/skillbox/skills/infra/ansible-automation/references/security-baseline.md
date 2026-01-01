# Ubuntu Security Baseline

Comprehensive hardening checklist for Ubuntu servers following CIS Benchmarks and production best practices.

## Hardening Layers

```
┌─────────────────────────────────────┐
│           Application               │  ← AppArmor profiles
├─────────────────────────────────────┤
│            Services                 │  ← Minimal, hardened
├─────────────────────────────────────┤
│          Filesystem                 │  ← Permissions, mount opts
├─────────────────────────────────────┤
│            Network                  │  ← Firewall, sysctl
├─────────────────────────────────────┤
│             Access                  │  ← SSH, users, sudo
├─────────────────────────────────────┤
│             Kernel                  │  ← Updates, sysctl
└─────────────────────────────────────┘
```

## Implementation Checklist

### 1. Access Control

| Item | Status | Priority |
|------|--------|----------|
| Create non-root admin user | Required | Critical |
| Configure SSH key-only authentication | Required | Critical |
| Disable root SSH login | Required | Critical |
| Configure sudo with NOPASSWD for deploy | Required | High |
| Install fail2ban or sshguard | Required | High |
| Set password policies | Recommended | Medium |
| Configure user session timeout | Recommended | Medium |

```yaml
# roles/baseline/tasks/users.yml
---
- name: Create deploy user
  ansible.builtin.user:
    name: "{{ baseline__deploy_user }}"
    groups: sudo
    shell: /bin/bash
    create_home: true
    state: present

- name: Set authorized keys for deploy user
  ansible.posix.authorized_key:
    user: "{{ baseline__deploy_user }}"
    key: "{{ item }}"
    state: present
  loop: "{{ baseline__ssh_public_keys }}"
  no_log: true

- name: Configure passwordless sudo
  ansible.builtin.copy:
    content: "{{ baseline__deploy_user }} ALL=(ALL) NOPASSWD:ALL"
    dest: "/etc/sudoers.d/{{ baseline__deploy_user }}"
    mode: '0440'
    validate: 'visudo -cf %s'
```

### 2. SSH Hardening

See `references/ssh-hardening.md` for detailed configuration.

Key requirements:
- Disable password authentication
- Disable root login
- Use strong ciphers/MACs
- Limit authentication attempts
- Configure fail2ban

### 3. Firewall Configuration

See `references/firewall-config.md` for detailed patterns.

Key requirements:
- Default deny incoming
- Whitelist only required ports
- Rate limit SSH connections
- Log denied connections

### 4. Automatic Security Updates

```yaml
# roles/baseline/tasks/updates.yml
---
- name: Install unattended-upgrades
  ansible.builtin.apt:
    name:
      - unattended-upgrades
      - apt-listchanges
    state: present
    update_cache: true

- name: Configure unattended-upgrades
  ansible.builtin.template:
    src: 50unattended-upgrades.j2
    dest: /etc/apt/apt.conf.d/50unattended-upgrades
    mode: '0644'

- name: Enable automatic updates
  ansible.builtin.template:
    src: 20auto-upgrades.j2
    dest: /etc/apt/apt.conf.d/20auto-upgrades
    mode: '0644'
```

Template for unattended-upgrades:

```jinja2
// templates/50unattended-upgrades.j2
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}";
    "${distro_id}:${distro_codename}-security";
    "${distro_id}ESMApps:${distro_codename}-apps-security";
    "${distro_id}ESM:${distro_codename}-infra-security";
};

Unattended-Upgrade::Package-Blacklist {
    // Packages that shouldn't be auto-upgraded
};

Unattended-Upgrade::DevRelease "false";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "{{ baseline__auto_reboot | lower }}";
Unattended-Upgrade::Automatic-Reboot-Time "{{ baseline__reboot_time | default('02:00') }}";

// Mail notifications
{% if baseline__updates_email is defined %}
Unattended-Upgrade::Mail "{{ baseline__updates_email }}";
Unattended-Upgrade::MailReport "on-change";
{% endif %}
```

### 5. AppArmor

```yaml
# roles/baseline/tasks/apparmor.yml
---
- name: Ensure AppArmor is installed
  ansible.builtin.apt:
    name:
      - apparmor
      - apparmor-utils
    state: present

- name: Ensure AppArmor is enabled
  ansible.builtin.systemd:
    name: apparmor
    enabled: true
    state: started

- name: Check AppArmor status
  ansible.builtin.command: aa-status
  register: apparmor_status
  changed_when: false

- name: Enforce profiles (except allowed exceptions)
  ansible.builtin.command: "aa-enforce /etc/apparmor.d/{{ item }}"
  loop: "{{ baseline__apparmor_enforce_profiles }}"
  when: item not in baseline__apparmor_exceptions
  changed_when: false
```

### 6. Kernel Hardening (sysctl)

```yaml
# roles/baseline/tasks/sysctl.yml
---
- name: Apply security sysctl settings
  ansible.posix.sysctl:
    name: "{{ item.key }}"
    value: "{{ item.value }}"
    sysctl_file: /etc/sysctl.d/99-security.conf
    reload: true
  loop: "{{ baseline__sysctl_settings | dict2items }}"

# defaults/main.yml
baseline__sysctl_settings:
  # Network security
  net.ipv4.conf.all.accept_redirects: 0
  net.ipv4.conf.default.accept_redirects: 0
  net.ipv6.conf.all.accept_redirects: 0
  net.ipv4.conf.all.send_redirects: 0
  net.ipv4.conf.all.accept_source_route: 0
  net.ipv6.conf.all.accept_source_route: 0
  net.ipv4.conf.all.log_martians: 1

  # TCP hardening
  net.ipv4.tcp_syncookies: 1
  net.ipv4.tcp_max_syn_backlog: 2048
  net.ipv4.tcp_synack_retries: 2

  # IP spoofing protection
  net.ipv4.conf.all.rp_filter: 1
  net.ipv4.conf.default.rp_filter: 1

  # Ignore ICMP broadcast requests
  net.ipv4.icmp_echo_ignore_broadcasts: 1

  # Disable IPv6 if not needed
  # net.ipv6.conf.all.disable_ipv6: 1
  # net.ipv6.conf.default.disable_ipv6: 1

  # Kernel hardening
  kernel.randomize_va_space: 2
  kernel.kptr_restrict: 2
  kernel.dmesg_restrict: 1
  kernel.perf_event_paranoid: 3
  kernel.yama.ptrace_scope: 2

  # Filesystem protection
  fs.protected_hardlinks: 1
  fs.protected_symlinks: 1
  fs.suid_dumpable: 0
```

### 7. Filesystem Security

```yaml
# roles/baseline/tasks/filesystem.yml
---
- name: Set secure permissions on critical directories
  ansible.builtin.file:
    path: "{{ item.path }}"
    mode: "{{ item.mode }}"
    owner: root
    group: root
  loop:
    - { path: '/etc/crontab', mode: '0600' }
    - { path: '/etc/cron.hourly', mode: '0700' }
    - { path: '/etc/cron.daily', mode: '0700' }
    - { path: '/etc/cron.weekly', mode: '0700' }
    - { path: '/etc/cron.monthly', mode: '0700' }
    - { path: '/etc/cron.d', mode: '0700' }

- name: Ensure /tmp is mounted with noexec,nosuid,nodev
  ansible.posix.mount:
    path: /tmp
    src: tmpfs
    fstype: tmpfs
    opts: "defaults,noexec,nosuid,nodev,size={{ baseline__tmp_size | default('2G') }}"
    state: mounted
  when: baseline__secure_tmp | default(true)
```

### 8. Time Synchronization

```yaml
# roles/baseline/tasks/ntp.yml
---
- name: Install chrony
  ansible.builtin.apt:
    name: chrony
    state: present

- name: Configure chrony
  ansible.builtin.template:
    src: chrony.conf.j2
    dest: /etc/chrony/chrony.conf
    mode: '0644'
  notify: Restart chrony

- name: Ensure chrony is running
  ansible.builtin.systemd:
    name: chrony
    enabled: true
    state: started

# handlers/main.yml
- name: Restart chrony
  ansible.builtin.systemd:
    name: chrony
    state: restarted
```

### 9. Audit Logging

```yaml
# roles/baseline/tasks/audit.yml
---
- name: Install auditd
  ansible.builtin.apt:
    name:
      - auditd
      - audispd-plugins
    state: present

- name: Configure audit rules
  ansible.builtin.template:
    src: audit.rules.j2
    dest: /etc/audit/rules.d/hardening.rules
    mode: '0640'
  notify: Restart auditd

- name: Ensure auditd is running
  ansible.builtin.systemd:
    name: auditd
    enabled: true
    state: started
```

Audit rules template:

```jinja2
# templates/audit.rules.j2
# Log all commands run by root
-a always,exit -F arch=b64 -F euid=0 -S execve -k root_commands

# Monitor SSH configuration
-w /etc/ssh/sshd_config -p wa -k sshd_config

# Monitor user/group changes
-w /etc/passwd -p wa -k identity
-w /etc/group -p wa -k identity
-w /etc/shadow -p wa -k identity
-w /etc/sudoers -p wa -k identity
-w /etc/sudoers.d/ -p wa -k identity

# Monitor network configuration
-w /etc/hosts -p wa -k network_config
-w /etc/network/ -p wa -k network_config
-w /etc/netplan/ -p wa -k network_config

# Monitor cron
-w /etc/crontab -p wa -k cron
-w /etc/cron.d/ -p wa -k cron

# Monitor time changes
-a always,exit -F arch=b64 -S adjtimex -S settimeofday -k time_change
-w /etc/localtime -p wa -k time_change
```

### 10. Remove Unnecessary Services

```yaml
# roles/baseline/tasks/services.yml
---
- name: Disable unnecessary services
  ansible.builtin.systemd:
    name: "{{ item }}"
    enabled: false
    state: stopped
  loop: "{{ baseline__disabled_services }}"
  failed_when: false

# defaults/main.yml
baseline__disabled_services:
  - avahi-daemon
  - cups
  - isc-dhcp-server
  - rpcbind
  - rsync
  - snapd  # Optional: disable if not using snaps
```

## CIS Benchmark Reference

For Ubuntu 24.04, use Ubuntu Security Guide (USG) for automated compliance:

```bash
# Install USG
sudo apt install ubuntu-security-guide

# Audit current state
sudo usg audit cis_level1_server

# Apply hardening profile
sudo usg fix cis_level1_server

# Generate compliance report
sudo usg audit cis_level1_server --output report.html
```

### CIS Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| Level 1 - Server | Basic hardening, minimal impact | Production servers |
| Level 2 - Server | Defense in depth, more restrictive | High-security environments |
| STIG | DoD security requirements | Government/compliance |

### Key CIS Controls

1. **Filesystem Configuration** - Separate partitions, mount options
2. **Software Updates** - Automatic security updates
3. **Filesystem Integrity** - AIDE for intrusion detection
4. **Secure Boot** - Boot settings, GRUB password
5. **Process Hardening** - Core dumps, ASLR
6. **Mandatory Access Control** - AppArmor enabled
7. **Warning Banners** - Login banners for legal notice
8. **inetd Services** - Disable xinetd/inetd
9. **Special Purpose Services** - Disable unused daemons
10. **Service Clients** - Remove unnecessary clients

## Role Execution Order

Apply hardening in correct order to avoid locking yourself out:

```yaml
# roles/baseline/tasks/main.yml
---
# 1. First ensure access is configured
- name: Configure users
  ansible.builtin.import_tasks: users.yml
  tags: [users, access]

# 2. SSH hardening (keep session alive during changes)
- name: Configure SSH
  ansible.builtin.import_tasks: ssh.yml
  tags: [ssh, access]

# 3. Firewall (ensure SSH allowed before enabling)
- name: Configure firewall
  ansible.builtin.import_tasks: firewall.yml
  tags: [firewall, network]

# 4. Rest of hardening
- name: Configure updates
  ansible.builtin.import_tasks: updates.yml
  tags: [updates]

- name: Configure sysctl
  ansible.builtin.import_tasks: sysctl.yml
  tags: [sysctl, kernel]

- name: Configure AppArmor
  ansible.builtin.import_tasks: apparmor.yml
  tags: [apparmor, mac]

- name: Configure filesystem
  ansible.builtin.import_tasks: filesystem.yml
  tags: [filesystem]

- name: Configure time sync
  ansible.builtin.import_tasks: ntp.yml
  tags: [ntp, time]

- name: Configure audit
  ansible.builtin.import_tasks: audit.yml
  tags: [audit, logging]

- name: Remove unnecessary services
  ansible.builtin.import_tasks: services.yml
  tags: [services]

# 5. Install fail2ban last (requires SSH config)
- name: Configure fail2ban
  ansible.builtin.import_tasks: fail2ban.yml
  tags: [fail2ban, security]
```

## Testing Hardening

### Local Testing

```bash
# Check mode first
ansible-playbook playbooks/baseline.yml --check --diff

# Run on single host
ansible-playbook playbooks/baseline.yml --limit server1

# Run specific tags
ansible-playbook playbooks/baseline.yml --tags ssh,firewall
```

### Verification Tasks

```yaml
# roles/baseline/tasks/verify.yml
---
- name: Verify SSH hardening
  ansible.builtin.command: sshd -T
  register: sshd_config
  changed_when: false

- name: Assert password auth disabled
  ansible.builtin.assert:
    that:
      - "'passwordauthentication no' in sshd_config.stdout | lower"
    fail_msg: "Password authentication is still enabled!"

- name: Verify firewall is active
  ansible.builtin.command: ufw status
  register: ufw_status
  changed_when: false

- name: Assert firewall is active
  ansible.builtin.assert:
    that:
      - "'Status: active' in ufw_status.stdout"
    fail_msg: "UFW is not active!"
```

## Recovery Plan

Always maintain break-glass access:

1. **Console access** - Cloud provider console, ILO/IPMI
2. **Secondary admin key** - Stored securely offline
3. **Documented recovery** - Step-by-step restore procedure
4. **Regular backups** - Test restoration periodically
5. **Rollback plan** - Keep previous configs for 24h
