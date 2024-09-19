#!/bin/bash

# Ensure the /var/www directory exists
mkdir -p /var/www

# Create the app directory if it doesn't exist
mkdir -p /var/www/{{.AppName}}

# Change to the app directory
cd /var/www/{{.AppName}}

# Check if the directory is empty
if [ -z "$(ls -A)" ]; then
    # If empty, clone the repository
    git clone {{.Repository}} .
else
    # If not empty, pull the latest changes
    git pull
fi

# Install dependencies
npm install

# Build the Next.js application
npm run build

# Restart the PM2 process (assuming PM2 is used for process management)
# pm2 restart {{.AppName}}

# Set appropriate permissions
chown -R www-data:www-data /var/www/{{.AppName}}
chmod -R 755 /var/www/{{.AppName}}

echo "Deployment of {{.AppName}} completed successfully!"
