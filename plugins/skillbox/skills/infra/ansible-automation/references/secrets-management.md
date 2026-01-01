# Secrets Management

Secure handling of sensitive data in Ansible projects.

## Principles

1. **Never store secrets in plain text** - Even in private repos
2. **Encrypt at rest** - Vault or external secret manager
3. **Minimal exposure** - `no_log: true` on sensitive tasks
4. **Rotate regularly** - Automated key/password rotation
5. **Audit access** - Track who accessed what and when

## Ansible Vault

### Basic Usage

```bash
# Create encrypted file
ansible-vault create secrets.yml

# Encrypt existing file
ansible-vault encrypt vars/secrets.yml

# Decrypt for editing
ansible-vault edit vars/secrets.yml

# View encrypted file
ansible-vault view vars/secrets.yml

# Decrypt file
ansible-vault decrypt vars/secrets.yml

# Change password
ansible-vault rekey vars/secrets.yml

# Encrypt string
ansible-vault encrypt_string 'super_secret' --name 'db_password'
```

### Vault Password Options

```bash
# Prompt for password
ansible-playbook site.yml --ask-vault-pass

# Use password file
ansible-playbook site.yml --vault-password-file=.vault_pass

# Use script (for CI/CD)
ansible-playbook site.yml --vault-password-file=./get_vault_pass.sh

# Multiple vault IDs (different passwords per environment)
ansible-playbook site.yml \
  --vault-id dev@.vault_pass_dev \
  --vault-id prod@.vault_pass_prod
```

### Password File

```bash
# Create password file
echo 'your_vault_password' > .vault_pass
chmod 600 .vault_pass

# Add to .gitignore
echo '.vault_pass*' >> .gitignore
```

### Encrypted Variables Pattern

```yaml
# group_vars/all/vault.yml (encrypted)
---
vault_db_password: "<your-db-password>"
vault_api_token: "<your-api-token>"
vault_ssh_key: |
  <your-ssh-private-key-content>

# group_vars/all/vars.yml (plain text, references vault)
---
db_password: "{{ vault_db_password }}"
api_token: "{{ vault_api_token }}"
ssh_key: "{{ vault_ssh_key }}"
```

### Inline Encrypted Strings

```yaml
# Encrypt a single value
# ansible-vault encrypt_string 'my_password' --name 'db_password'

db_password: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  61626364656667686970716B6C6D6E6F70717273747576777879...

# With vault ID
# ansible-vault encrypt_string 'my_password' --vault-id prod@.vault_pass_prod --name 'db_password'

db_password: !vault |
  $ANSIBLE_VAULT;1.2;AES256;prod
  61626364656667686970716B6C6D6E6F70717273747576777879...
```

## no_log Pattern

Always hide sensitive data from logs:

```yaml
# BAD: Password visible in logs
- name: Create database user
  community.postgresql.postgresql_user:
    name: app
    password: "{{ db_password }}"

# GOOD: Password hidden
- name: Create database user
  community.postgresql.postgresql_user:
    name: app
    password: "{{ db_password }}"
  no_log: true

# GOOD: Conditional logging (show in debug mode)
- name: Create database user
  community.postgresql.postgresql_user:
    name: app
    password: "{{ db_password }}"
  no_log: "{{ not (ansible_verbosity >= 3) }}"
```

### Handling Registered Variables

```yaml
# Secret might leak through registered output
- name: Get API token
  ansible.builtin.uri:
    url: https://api.example.com/auth
    method: POST
    body:
      username: "{{ api_user }}"
      password: "{{ api_password }}"
    body_format: json
  register: auth_response
  no_log: true

# Safe to use the token
- name: Use API
  ansible.builtin.uri:
    url: https://api.example.com/data
    headers:
      Authorization: "Bearer {{ auth_response.json.token }}"
  no_log: true
```

## External Secret Managers

### HashiCorp Vault Integration

```yaml
# requirements.yml
collections:
  - name: community.hashi_vault
    version: ">=6.0.0,<7.0.0"
```

```yaml
# Lookup secret from HashiCorp Vault
- name: Get database password
  ansible.builtin.set_fact:
    db_password: "{{ lookup('community.hashi_vault.hashi_vault',
      'secret/data/myapp/database:password',
      url='https://vault.example.com:8200',
      token=lookup('env', 'VAULT_TOKEN')) }}"
  no_log: true

# Using environment variables
- name: Get secret using env auth
  ansible.builtin.set_fact:
    api_key: "{{ lookup('community.hashi_vault.vault_kv2_get',
      'myapp/api',
      engine_mount_point='secret') }}"
  environment:
    VAULT_ADDR: "https://vault.example.com:8200"
    VAULT_TOKEN: "{{ lookup('env', 'VAULT_TOKEN') }}"
  no_log: true
```

### AWS Secrets Manager

```yaml
# requirements.yml
collections:
  - name: amazon.aws
    version: ">=7.0.0,<8.0.0"
```

```yaml
# Lookup from AWS Secrets Manager
- name: Get database credentials
  ansible.builtin.set_fact:
    db_creds: "{{ lookup('amazon.aws.aws_secret',
      'myapp/database',
      region='us-east-1') | from_json }}"
  no_log: true

- name: Configure database
  ansible.builtin.template:
    src: database.conf.j2
    dest: /etc/app/database.conf
  vars:
    db_host: "{{ db_creds.host }}"
    db_user: "{{ db_creds.username }}"
    db_pass: "{{ db_creds.password }}"
  no_log: true
```

### Azure Key Vault

```yaml
# requirements.yml
collections:
  - name: azure.azcollection
    version: ">=2.0.0,<3.0.0"
```

```yaml
- name: Get secret from Azure Key Vault
  azure.azcollection.azure_rm_keyvaultsecret_info:
    vault_uri: "https://myvault.vault.azure.net"
    name: "db-password"
  register: vault_secret
  no_log: true

- name: Use the secret
  ansible.builtin.set_fact:
    db_password: "{{ vault_secret.secrets[0].secret }}"
  no_log: true
```

### GCP Secret Manager

```yaml
# requirements.yml
collections:
  - name: google.cloud
    version: ">=1.3.0,<2.0.0"
```

```yaml
- name: Get secret from GCP
  google.cloud.gcp_secretmanager_secret_version_info:
    project: my-project
    secret: db-password
    version: latest
  register: secret_version
  no_log: true

- name: Decode secret
  ansible.builtin.set_fact:
    db_password: "{{ secret_version.payload.data | b64decode }}"
  no_log: true
```

## CI/CD Secrets

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Create vault password file
        run: echo "${{ secrets.VAULT_PASSWORD }}" > .vault_pass
        shell: bash

      - name: Run Ansible
        run: |
          ansible-playbook site.yml \
            --vault-password-file=.vault_pass
        env:
          ANSIBLE_HOST_KEY_CHECKING: "false"

      - name: Cleanup
        if: always()
        run: rm -f .vault_pass
```

### GitLab CI

```yaml
# .gitlab-ci.yml
deploy:
  stage: deploy
  script:
    - echo "$VAULT_PASSWORD" > .vault_pass
    - ansible-playbook site.yml --vault-password-file=.vault_pass
  after_script:
    - rm -f .vault_pass
  variables:
    ANSIBLE_HOST_KEY_CHECKING: "false"
```

### Vault Password Script

```bash
#!/bin/bash
# get_vault_pass.sh - Fetch vault password from external source

# From environment variable
if [ -n "$ANSIBLE_VAULT_PASSWORD" ]; then
    echo "$ANSIBLE_VAULT_PASSWORD"
    exit 0
fi

# From AWS Secrets Manager
if command -v aws &> /dev/null; then
    aws secretsmanager get-secret-value \
        --secret-id ansible/vault-password \
        --query SecretString \
        --output text
    exit 0
fi

# From file (fallback)
if [ -f ~/.vault_pass ]; then
    cat ~/.vault_pass
    exit 0
fi

echo "No vault password source found" >&2
exit 1
```

## Best Practices

### File Organization

```
group_vars/
├── all/
│   ├── vars.yml         # Plain text, references vault_* vars
│   └── vault.yml        # Encrypted, contains vault_* vars
├── prod/
│   ├── vars.yml
│   └── vault.yml        # Different secrets for prod
└── dev/
    ├── vars.yml
    └── vault.yml        # Different secrets for dev
```

### Naming Convention

```yaml
# Encrypted file (vault.yml)
vault_db_password: "actual_password"
vault_api_key: "actual_key"

# Reference file (vars.yml)
db_password: "{{ vault_db_password }}"
api_key: "{{ vault_api_key }}"
```

### Role-specific Secrets

```yaml
# roles/database/defaults/main.yml
database__admin_password: ""  # Must be set in inventory

# inventories/prod/group_vars/databases/vault.yml
vault_database_admin_password: "super_secret"

# inventories/prod/group_vars/databases/vars.yml
database__admin_password: "{{ vault_database_admin_password }}"
```

## Security Checklist

| Item | Status | Notes |
|------|--------|-------|
| All secrets encrypted with Vault | Required | No plain text secrets in repo |
| no_log on sensitive tasks | Required | Prevent log exposure |
| Vault password not in repo | Required | Use file, env var, or script |
| Different secrets per env | Recommended | Separate dev/prod credentials |
| External secret manager for prod | Recommended | HashiCorp Vault, AWS SM, etc. |
| Regular secret rotation | Recommended | Automated where possible |
| Audit logging enabled | Recommended | Track secret access |
| .vault_pass in .gitignore | Required | Never commit password files |

## Troubleshooting

### Common Issues

```bash
# Decrypt error - wrong password
ERROR! Decryption failed

# Check vault file format
head -1 vars/vault.yml
# Should show: $ANSIBLE_VAULT;1.1;AES256

# Verify password file
cat .vault_pass | xxd  # Check for hidden characters

# Test decryption
ansible-vault view vars/vault.yml --vault-password-file=.vault_pass
```

### Recovering from Mistakes

```bash
# Committed secret by accident
# 1. Rotate the secret immediately
# 2. Remove from git history
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch path/to/secrets.yml' \
  --prune-empty --tag-name-filter cat -- --all

# 3. Force push (dangerous!)
git push origin --force --all
git push origin --force --tags

# Better: Use git-filter-repo
pip install git-filter-repo
git filter-repo --path path/to/secrets.yml --invert-paths
```

## Summary

1. **Always use Vault** for secrets in Ansible
2. **no_log: true** on all sensitive tasks
3. **External managers** for production environments
4. **Separate vault per environment** using vault IDs
5. **Script-based passwords** for CI/CD integration
6. **Regular rotation** with automation where possible
7. **Audit trail** for compliance requirements
