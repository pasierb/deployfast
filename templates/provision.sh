#!/bin/bash

setup_deployer_user() {
    # Create deployer user if it doesn't exist
    if ! id "deployer" &>/dev/null; then
        useradd -m -s /bin/bash -p '*' deployer
        echo "Deployer user created."
    else
        echo "Deployer user already exists."
    fi

    # Add deployer user to www-data group if not already a member
    if ! groups deployer | grep -q "\bwww-data\b"; then
        usermod -aG www-data deployer
        echo "Added deployer to www-data group."
    else
        echo "Deployer is already in www-data group."
    fi

    # Setup SSH for deployer
    setup_ssh_for_deployer

    # Disable password authentication for the deployer user
    passwd -l deployer

    echo "Deployer user setup complete (SSH access only)!"
}

setup_ssh_for_deployer() {
    # Create .ssh directory for the deployer user if it doesn't exist
    if [ ! -d "/home/deployer/.ssh" ]; then
        mkdir -p /home/deployer/.ssh
        chmod 700 /home/deployer/.ssh
        echo "Created .ssh directory for deployer."
    else
        echo ".ssh directory already exists for deployer."
    fi

    # Create authorized_keys file if it doesn't exist
    if [ ! -f "/home/deployer/.ssh/authorized_keys" ]; then
        touch /home/deployer/.ssh/authorized_keys
        chmod 600 /home/deployer/.ssh/authorized_keys
        echo "Created authorized_keys file for deployer."
    else
        echo "authorized_keys file already exists for deployer."
    fi

    # Set proper ownership
    chown -R deployer:deployer /home/deployer/.ssh
}

update_system_packages() {
    apt-get update
    apt-get upgrade -y
    echo "System packages updated."
}

install_and_configure_nginx() {
    # Install Nginx web server if not already installed
    if ! command -v nginx &> /dev/null; then
        apt-get install nginx -y
        echo "Nginx installed."
    else
        echo "Nginx is already installed."
    fi

    # Ensure Nginx service is started
    service nginx start
    echo "Nginx started using service command."

    # Set proper permissions for web root
    chown -R www-data:www-data /var/www/html
    chmod -R 755 /var/www/html

    # Restart Nginx to apply changes
    service nginx restart
    echo "Nginx restarted using service command."

    echo "Nginx web server provisioning complete!"
}

install_nodejs() {
    # Remove any existing installations
    apt-get purge nodejs npm -y
    apt-get autoremove -y

    # Install curl if not already installed
    if ! command -v curl &> /dev/null; then
        apt-get update
        apt-get install -y curl
    fi

    # Download and execute the NodeSource setup script for Node.js 20.x
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -

    # Install Node.js
    apt-get install -y nodejs

    # Verify Node.js installation
    node_version=$(node -v)
    npm_version=$(npm -v)
    echo "Node.js version: $node_version"
    echo "npm version: $npm_version"

    echo "Node.js installation complete!"

    # Install PM2 globally
    npm install -g pm2
    pm2_version=$(pm2 -v)
    echo "PM2 installed globally. Version: $pm2_version"
}

install_git() {
    if ! command -v git &> /dev/null; then
        apt-get update
        apt-get install -y git
        echo "Git installed successfully."
    else
        echo "Git is already installed."
    fi
    
    git_version=$(git --version)
    echo "Git version: $git_version"
}

# Main execution
setup_deployer_user
update_system_packages
install_and_configure_nginx
install_nodejs
install_git

