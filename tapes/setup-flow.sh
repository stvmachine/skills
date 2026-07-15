# 1. Clone (SSH — repo is private) and install
git clone git@github.com:nimblic/medtasker-skills.git
cd medtasker-skills && ./scripts/install.sh

# 2. Store your API tokens (encrypted at rest via dotenvx)
medtasker-skills env setup

# 3. Launch Claude — dotenvx decrypts credentials in-process,
#    Claude sees them as env vars and wires up Jira, GitHub, Figma...
dotenvx run -f ~/.medtasker-skills/.env -- claude
