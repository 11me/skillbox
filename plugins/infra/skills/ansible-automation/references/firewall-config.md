# Firewall Configuration

UFW and nftables patterns for Ubuntu server hardening.

## Core Principle: Default Deny

```
┌─────────────────────────────────────────┐
│              INCOMING                    │
│  ┌───────────────────────────────────┐  │
│  │         DEFAULT: DENY             │  │
│  │  ┌─────────┐  ┌─────────┐        │  │
│  │  │ SSH:22  │  │ HTTP:80 │  ...   │  │
│  │  │ ALLOW   │  │ ALLOW   │        │  │
│  │  └─────────┘  └─────────┘        │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

## UFW Implementation

### Basic Setup

```yaml
# roles/baseline/tasks/firewall.yml
---
- name: Install UFW
  ansible.builtin.apt:
    name: ufw
    state: present
    update_cache: true

- name: Reset UFW to defaults
  community.general.ufw:
    state: reset
  when: baseline__firewall_reset | default(false)

- name: Set default policies
  community.general.ufw:
    direction: "{{ item.direction }}"
    policy: "{{ item.policy }}"
  loop:
    - { direction: incoming, policy: deny }
    - { direction: outgoing, policy: allow }
    - { direction: routed, policy: deny }

- name: Allow SSH (critical - do this before enabling!)
  community.general.ufw:
    rule: allow
    port: "{{ baseline__ssh_port | default(22) }}"
    proto: tcp
    comment: "SSH access"

- name: Allow configured ports
  community.general.ufw:
    rule: allow
    port: "{{ item.port }}"
    proto: "{{ item.proto | default('tcp') }}"
    from_ip: "{{ item.from | default('any') }}"
    comment: "{{ item.comment | default(omit) }}"
  loop: "{{ baseline__firewall_allowed_ports }}"

- name: Enable UFW
  community.general.ufw:
    state: enabled
    logging: "{{ baseline__firewall_logging | default('on') }}"

- name: Enable UFW service
  ansible.builtin.systemd:
    name: ufw
    enabled: true
    state: started
```

### Role Variables

```yaml
# roles/baseline/defaults/main.yml
---
# Firewall settings
baseline__firewall_reset: false
baseline__firewall_logging: "on"  # on, off, low, medium, high, full

# Ports to allow (beyond SSH which is always allowed)
baseline__firewall_allowed_ports:
  - port: 80
    proto: tcp
    comment: "HTTP"
  - port: 443
    proto: tcp
    comment: "HTTPS"

# Rate limiting for SSH
baseline__firewall_rate_limit_ssh: true

# Restrict SSH to specific IPs (if known)
# baseline__firewall_ssh_allowed_ips:
#   - 10.0.0.0/8
#   - 192.168.1.0/24
```

### Rate Limiting

```yaml
# Rate limit SSH to prevent brute force
- name: Rate limit SSH connections
  community.general.ufw:
    rule: limit
    port: "{{ baseline__ssh_port | default(22) }}"
    proto: tcp
    comment: "Rate limit SSH"
  when: baseline__firewall_rate_limit_ssh | default(true)
```

### IP-based Restrictions

```yaml
# Restrict SSH to specific IPs
- name: Allow SSH from specific IPs only
  community.general.ufw:
    rule: allow
    port: "{{ baseline__ssh_port | default(22) }}"
    proto: tcp
    from_ip: "{{ item }}"
    comment: "SSH from {{ item }}"
  loop: "{{ baseline__firewall_ssh_allowed_ips }}"
  when: baseline__firewall_ssh_allowed_ips is defined

# Block SSH from everywhere else if restricted
- name: Block SSH from all other IPs
  community.general.ufw:
    rule: deny
    port: "{{ baseline__ssh_port | default(22) }}"
    proto: tcp
    comment: "Block SSH from other IPs"
  when: baseline__firewall_ssh_allowed_ips is defined
```

### Application Profiles

```yaml
# Use UFW application profiles
- name: Allow application profiles
  community.general.ufw:
    rule: allow
    name: "{{ item }}"
  loop: "{{ baseline__firewall_app_profiles | default([]) }}"

# Example profiles: OpenSSH, Nginx Full, Apache Full
```

## nftables Implementation

For more advanced setups or when UFW is not suitable:

```yaml
# roles/baseline/tasks/nftables.yml
---
- name: Install nftables
  ansible.builtin.apt:
    name: nftables
    state: present

- name: Disable UFW if using nftables
  ansible.builtin.systemd:
    name: ufw
    enabled: false
    state: stopped
  failed_when: false

- name: Configure nftables ruleset
  ansible.builtin.template:
    src: nftables.conf.j2
    dest: /etc/nftables.conf
    mode: '0600'
    validate: 'nft -c -f %s'
  notify: Reload nftables

- name: Enable nftables service
  ansible.builtin.systemd:
    name: nftables
    enabled: true
    state: started
```

nftables template:

```jinja2
#!/usr/sbin/nft -f
# templates/nftables.conf.j2

flush ruleset

table inet filter {
    # Connection tracking
    chain input {
        type filter hook input priority 0; policy drop;

        # Allow established connections
        ct state established,related accept

        # Drop invalid packets
        ct state invalid drop

        # Allow loopback
        iifname lo accept

        # Allow ICMP (ping)
        ip protocol icmp accept
        ip6 nexthdr icmpv6 accept

        # SSH (rate limited)
        tcp dport {{ baseline__ssh_port | default(22) }} ct state new limit rate 4/minute accept comment "SSH rate limit"

{% for port in baseline__firewall_allowed_ports %}
        # {{ port.comment | default('Port ' + port.port | string) }}
{% if port.from is defined and port.from != 'any' %}
        ip saddr {{ port.from }} {{ port.proto | default('tcp') }} dport {{ port.port }} accept
{% else %}
        {{ port.proto | default('tcp') }} dport {{ port.port }} accept
{% endif %}
{% endfor %}

        # Log and drop everything else
        log prefix "nftables-drop: " level info
        counter drop
    }

    chain forward {
        type filter hook forward priority 0; policy drop;

        # Allow established forwarded connections
        ct state established,related accept

{% if baseline__firewall_forward_rules is defined %}
{% for rule in baseline__firewall_forward_rules %}
        # {{ rule.comment | default('Forward rule') }}
        {{ rule.rule }}
{% endfor %}
{% endif %}
    }

    chain output {
        type filter hook output priority 0; policy accept;

        # Allow all outgoing by default
        ct state established,related accept
    }
}

{% if baseline__firewall_nat_rules is defined %}
table ip nat {
    chain prerouting {
        type nat hook prerouting priority -100; policy accept;
{% for rule in baseline__firewall_nat_rules.prerouting | default([]) %}
        {{ rule }}
{% endfor %}
    }

    chain postrouting {
        type nat hook postrouting priority 100; policy accept;
{% for rule in baseline__firewall_nat_rules.postrouting | default([]) %}
        {{ rule }}
{% endfor %}
    }
}
{% endif %}
```

### nftables Handler

```yaml
# handlers/main.yml
- name: Reload nftables
  ansible.builtin.systemd:
    name: nftables
    state: restarted
```

## Port Reference

Common ports to consider:

| Port | Service | Notes |
|------|---------|-------|
| 22 | SSH | Always allow, rate limit |
| 80 | HTTP | Redirect to HTTPS |
| 443 | HTTPS | Main web traffic |
| 25 | SMTP | Email sending (outbound only usually) |
| 587 | SMTP Submission | Email with auth |
| 993 | IMAPS | Email reading |
| 3306 | MySQL | Restrict to app servers |
| 5432 | PostgreSQL | Restrict to app servers |
| 6379 | Redis | Internal only |
| 9090 | Prometheus | Internal only |
| 9100 | Node Exporter | Internal only |

## Environment-Specific Rules

```yaml
# inventories/dev/group_vars/all.yml
baseline__firewall_allowed_ports:
  - port: 80
    proto: tcp
  - port: 443
    proto: tcp
  - port: 8080
    proto: tcp
    comment: "Dev debug port"

# inventories/prod/group_vars/all.yml
baseline__firewall_allowed_ports:
  - port: 80
    proto: tcp
  - port: 443
    proto: tcp
# No debug ports in prod!

baseline__firewall_ssh_allowed_ips:
  - 10.0.0.0/8  # VPN only
```

## Docker Considerations

Docker manipulates iptables directly, bypassing UFW:

```yaml
# Prevent Docker from bypassing firewall
- name: Configure Docker to respect iptables
  ansible.builtin.copy:
    content: |
      {
        "iptables": false
      }
    dest: /etc/docker/daemon.json
    mode: '0644'
  notify: Restart docker
  when: baseline__firewall_docker_fix | default(false)

# Alternative: use docker network with explicit port binding
# docker run -p 127.0.0.1:8080:80 myapp
```

## Verification

```yaml
# roles/baseline/tasks/verify_firewall.yml
---
- name: Check UFW status
  ansible.builtin.command: ufw status verbose
  register: ufw_status
  changed_when: false

- name: Assert UFW is active
  ansible.builtin.assert:
    that:
      - "'Status: active' in ufw_status.stdout"
    fail_msg: "UFW is not active!"

- name: Display firewall rules
  ansible.builtin.debug:
    var: ufw_status.stdout_lines

- name: Verify SSH port is reachable
  ansible.builtin.wait_for:
    port: "{{ baseline__ssh_port }}"
    host: "{{ inventory_hostname }}"
    timeout: 5
  delegate_to: localhost
  become: false
```

## Troubleshooting

### Common Commands

```bash
# UFW
sudo ufw status verbose
sudo ufw status numbered
sudo ufw show raw
sudo ufw allow 8080/tcp
sudo ufw delete 5  # by rule number
sudo ufw reset

# nftables
sudo nft list ruleset
sudo nft list table inet filter
sudo nft monitor  # live traffic

# Check listening ports
sudo ss -tlnp
sudo netstat -tlnp

# Test connectivity
nc -vz host port
telnet host port
```

### Emergency Recovery

If locked out due to firewall:

1. **Cloud Console** - Access via provider's console
2. **Rescue Mode** - Boot into rescue, mount disk, edit rules
3. **Out-of-band** - ILO, IPMI, serial console

```bash
# From rescue/console, disable firewall
sudo ufw disable
# or
sudo systemctl stop nftables

# Then fix rules and re-enable
sudo ufw enable
```

## Best Practices

1. **Always allow SSH first** before enabling firewall
2. **Test in check mode** before applying
3. **Keep console access** as backup
4. **Rate limit SSH** to prevent brute force
5. **Log denied packets** for troubleshooting
6. **Minimize open ports** - only what's needed
7. **Use application profiles** when available
8. **Document all rules** with comments
9. **Regular audits** - review open ports periodically
10. **Separate concerns** - different rules per environment
