# SSH Hardening

Comprehensive SSH security configuration for Ubuntu servers.

## Security Principles

1. **Key-only authentication** - Disable password auth completely
2. **No root login** - Force user escalation with sudo
3. **Strong ciphers** - Disable weak algorithms
4. **Rate limiting** - Protect against brute force
5. **Access control** - Restrict who can connect

## Complete sshd_config Template

```jinja2
# templates/sshd_config.j2
# SSH Server Configuration - Hardened

# Network
Port {{ baseline__ssh_port | default(22) }}
AddressFamily {{ baseline__ssh_address_family | default('any') }}
ListenAddress {{ baseline__ssh_listen_address | default('0.0.0.0') }}

# Protocol
Protocol 2

# Host Keys
HostKey /etc/ssh/ssh_host_ed25519_key
HostKey /etc/ssh/ssh_host_rsa_key

# Ciphers and Algorithms (Strong only)
Ciphers chacha20-poly1305@openssh.com,aes256-gcm@openssh.com,aes128-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr
MACs hmac-sha2-512-etm@openssh.com,hmac-sha2-256-etm@openssh.com,hmac-sha2-512,hmac-sha2-256
KexAlgorithms sntrup761x25519-sha512@openssh.com,curve25519-sha256,curve25519-sha256@libssh.org,diffie-hellman-group18-sha512,diffie-hellman-group16-sha512
HostKeyAlgorithms ssh-ed25519,ssh-ed25519-cert-v01@openssh.com,rsa-sha2-512,rsa-sha2-256

# Logging
SyslogFacility AUTH
LogLevel VERBOSE

# Authentication
LoginGraceTime 30
PermitRootLogin no
StrictModes yes
MaxAuthTries {{ baseline__ssh_max_auth_tries | default(3) }}
MaxSessions {{ baseline__ssh_max_sessions | default(3) }}
MaxStartups 10:30:60

PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys

# Disable password auth completely
PasswordAuthentication no
PermitEmptyPasswords no
ChallengeResponseAuthentication no

# Disable unused auth methods
KerberosAuthentication no
GSSAPIAuthentication no
HostbasedAuthentication no
IgnoreRhosts yes

# Environment
PermitUserEnvironment no

# Disable forwarding (enable only if needed)
AllowAgentForwarding {{ baseline__ssh_allow_agent_forwarding | default('no') }}
AllowTcpForwarding {{ baseline__ssh_allow_tcp_forwarding | default('no') }}
X11Forwarding no
GatewayPorts no
PermitTunnel no

# Misc
PrintMotd no
PrintLastLog yes
TCPKeepAlive yes
ClientAliveInterval 300
ClientAliveCountMax 2
UseDNS no
PermitUserRC no

# Banner
Banner {{ baseline__ssh_banner | default('/etc/ssh/banner') }}

# Subsystems
Subsystem sftp /usr/lib/openssh/sftp-server -f AUTHPRIV -l INFO

{% if baseline__ssh_allow_users is defined and baseline__ssh_allow_users | length > 0 %}
# User restrictions
AllowUsers {{ baseline__ssh_allow_users | join(' ') }}
{% endif %}

{% if baseline__ssh_allow_groups is defined and baseline__ssh_allow_groups | length > 0 %}
AllowGroups {{ baseline__ssh_allow_groups | join(' ') }}
{% endif %}

{% if baseline__ssh_deny_users is defined and baseline__ssh_deny_users | length > 0 %}
DenyUsers {{ baseline__ssh_deny_users | join(' ') }}
{% endif %}

# Match block examples
{% if baseline__ssh_match_blocks is defined %}
{% for block in baseline__ssh_match_blocks %}
Match {{ block.type }} {{ block.pattern }}
{% for setting, value in block.settings.items() %}
    {{ setting }} {{ value }}
{% endfor %}

{% endfor %}
{% endif %}
```

## Ansible Implementation

```yaml
# roles/baseline/tasks/ssh.yml
---
- name: Generate SSH host keys if missing
  ansible.builtin.command:
    cmd: ssh-keygen -A
  args:
    creates: /etc/ssh/ssh_host_ed25519_key
  notify: Restart sshd

- name: Remove weak host keys
  ansible.builtin.file:
    path: "/etc/ssh/{{ item }}"
    state: absent
  loop:
    - ssh_host_dsa_key
    - ssh_host_dsa_key.pub
    - ssh_host_ecdsa_key
    - ssh_host_ecdsa_key.pub
  notify: Restart sshd

- name: Configure sshd
  ansible.builtin.template:
    src: sshd_config.j2
    dest: /etc/ssh/sshd_config
    owner: root
    group: root
    mode: '0600'
    validate: '/usr/sbin/sshd -t -f %s'
  notify: Restart sshd

- name: Create SSH banner
  ansible.builtin.copy:
    content: |
      **************************************************************************
      *                                                                        *
      * This system is for authorized use only.                                *
      * All connections and activity are logged and monitored.                 *
      * Unauthorized access is prohibited and will be prosecuted.              *
      *                                                                        *
      **************************************************************************
    dest: /etc/ssh/banner
    mode: '0644'

- name: Set correct permissions on .ssh directories
  ansible.builtin.file:
    path: "/home/{{ item }}/.ssh"
    state: directory
    owner: "{{ item }}"
    group: "{{ item }}"
    mode: '0700'
  loop: "{{ baseline__ssh_users | default([baseline__deploy_user]) }}"

- name: Set correct permissions on authorized_keys
  ansible.builtin.file:
    path: "/home/{{ item }}/.ssh/authorized_keys"
    owner: "{{ item }}"
    group: "{{ item }}"
    mode: '0600'
  loop: "{{ baseline__ssh_users | default([baseline__deploy_user]) }}"
  failed_when: false

# handlers/main.yml
- name: Restart sshd
  ansible.builtin.systemd:
    name: sshd
    state: restarted
```

## Role Variables

```yaml
# roles/baseline/defaults/main.yml
---
# SSH Configuration
baseline__ssh_port: 22
baseline__ssh_max_auth_tries: 3
baseline__ssh_max_sessions: 3
baseline__ssh_allow_agent_forwarding: "no"
baseline__ssh_allow_tcp_forwarding: "no"
baseline__ssh_banner: "/etc/ssh/banner"

# Access restrictions (optional)
# baseline__ssh_allow_users:
#   - deploy
#   - admin

baseline__ssh_allow_groups:
  - ssh-users

# Match blocks for conditional settings
baseline__ssh_match_blocks: []
# Example:
#   - type: User
#     pattern: sftp-user
#     settings:
#       ForceCommand: internal-sftp
#       ChrootDirectory: /var/sftp/%u
#       AllowTcpForwarding: "no"
```

## fail2ban Configuration

```yaml
# roles/baseline/tasks/fail2ban.yml
---
- name: Install fail2ban
  ansible.builtin.apt:
    name: fail2ban
    state: present
    update_cache: true

- name: Create fail2ban SSH jail
  ansible.builtin.template:
    src: jail.local.j2
    dest: /etc/fail2ban/jail.local
    mode: '0644'
  notify: Restart fail2ban

- name: Ensure fail2ban is running
  ansible.builtin.systemd:
    name: fail2ban
    enabled: true
    state: started
```

fail2ban jail template:

```jinja2
# templates/jail.local.j2
[DEFAULT]
# Ban duration (seconds, or -1 for permanent)
bantime = {{ baseline__fail2ban_bantime | default('1h') }}

# Time window for counting failures
findtime = {{ baseline__fail2ban_findtime | default('10m') }}

# Number of failures before ban
maxretry = {{ baseline__fail2ban_maxretry | default(5) }}

# Action to take
banaction = ufw
banaction_allports = ufw

# Ignore local networks
ignoreip = 127.0.0.1/8 ::1 {{ baseline__fail2ban_ignoreip | default('') }}

# Email notifications
{% if baseline__fail2ban_destemail is defined %}
destemail = {{ baseline__fail2ban_destemail }}
sender = fail2ban@{{ ansible_fqdn }}
mta = sendmail
action = %(action_mwl)s
{% endif %}

[sshd]
enabled = true
port = {{ baseline__ssh_port | default(22) }}
filter = sshd
logpath = /var/log/auth.log
maxretry = {{ baseline__fail2ban_ssh_maxretry | default(3) }}
bantime = {{ baseline__fail2ban_ssh_bantime | default('24h') }}
findtime = {{ baseline__fail2ban_ssh_findtime | default('10m') }}

{% if baseline__fail2ban_aggressive %}
# Aggressive mode: recidive jail for repeat offenders
[recidive]
enabled = true
logpath = /var/log/fail2ban.log
banaction = ufw
bantime = 1w
findtime = 1d
maxretry = 5
{% endif %}
```

## Testing SSH Configuration

```bash
# Validate config before restart
sudo sshd -t

# Test connection in verbose mode
ssh -v user@host

# Check active algorithms
ssh -Q cipher
ssh -Q mac
ssh -Q kex

# Verify settings
sudo sshd -T | grep -E "(passwordauthentication|permitrootlogin|pubkeyauthentication)"

# Check fail2ban status
sudo fail2ban-client status sshd
```

## Security Verification Tasks

```yaml
# roles/baseline/tasks/verify_ssh.yml
---
- name: Verify sshd configuration
  ansible.builtin.command: sshd -T
  register: sshd_config
  changed_when: false

- name: Assert security settings
  ansible.builtin.assert:
    that:
      - "'passwordauthentication no' in sshd_config.stdout | lower"
      - "'permitrootlogin no' in sshd_config.stdout | lower"
      - "'pubkeyauthentication yes' in sshd_config.stdout | lower"
      - "'permitemptypasswords no' in sshd_config.stdout | lower"
    fail_msg: "SSH security settings verification failed!"
    success_msg: "SSH is properly hardened"

- name: Verify fail2ban is active
  ansible.builtin.command: fail2ban-client status sshd
  register: fail2ban_status
  changed_when: false
  failed_when: "'is not running' in fail2ban_status.stderr | default('')"

- name: Check SSH port is open
  ansible.builtin.wait_for:
    port: "{{ baseline__ssh_port }}"
    timeout: 5
  delegate_to: localhost
  become: false
```

## Troubleshooting

### Common Issues

| Problem | Cause | Solution |
|---------|-------|----------|
| Connection refused | sshd not running or wrong port | Check `systemctl status sshd`, verify port |
| Permission denied | Key not authorized | Check `~/.ssh/authorized_keys` permissions (600) |
| Too many auth failures | fail2ban ban | `fail2ban-client set sshd unbanip <IP>` |
| No matching key exchange | Weak algorithms | Update client or add compatible algorithms |
| Host key verification failed | Key changed | Remove old key from `known_hosts` |

### Recovery Commands

```bash
# Unban IP
sudo fail2ban-client set sshd unbanip 192.168.1.100

# Temporarily allow password auth (emergency only)
sudo sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
sudo systemctl restart sshd
# Don't forget to revert!

# Check auth logs
sudo tail -f /var/log/auth.log

# Restart services
sudo systemctl restart sshd
sudo systemctl restart fail2ban
```

## Alternative: sshguard

For simpler setup or Alpine/musl systems:

```yaml
- name: Install sshguard
  ansible.builtin.apt:
    name: sshguard
    state: present

- name: Configure sshguard
  ansible.builtin.template:
    src: sshguard.conf.j2
    dest: /etc/sshguard/sshguard.conf
    mode: '0644'
  notify: Restart sshguard

- name: Ensure sshguard is running
  ansible.builtin.systemd:
    name: sshguard
    enabled: true
    state: started
```

## Best Practices Summary

1. **Always test changes** - Use `--check --diff` first
2. **Keep a backdoor** - Console access, second admin key
3. **Log everything** - VERBOSE logging for audits
4. **Rate limit** - fail2ban or firewall-based limiting
5. **Update regularly** - Keep OpenSSH current
6. **Monitor** - Alert on failed login attempts
7. **Rotate keys** - Periodically regenerate host keys
8. **Audit access** - Review authorized_keys regularly
