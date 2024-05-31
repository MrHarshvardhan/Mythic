#!/bin/bash

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to handle errors
handle_error() {
    local exit_code=$?
    local msg="$1"
    if [ $exit_code -ne 0 ]; then
        echo "Error: $msg"
        echo "Exit code: $exit_code"
        case $exit_code in
            1) echo "General error (miscellaneous errors, such as 'divide by zero' and other impermissible operations)";;
            2) echo "Misuse of shell builtins (according to Bash documentation)";;
            126) echo "Command invoked cannot execute";;
            127) echo "Command not found";;
            128) echo "Invalid argument to exit";;
            130) echo "Script terminated by Control-C";;
            137) echo "Script terminated by SIGKILL";;
            *) echo "An unknown error occurred";;
        esac
        exit $exit_code
    fi
}

# Function to install Docker
install_docker() {
    echo "Installing Docker..."
    sudo apt-get update || handle_error "Failed to update package list"
    sudo apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release || handle_error "Failed to install Docker dependencies"

    curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg || handle_error "Failed to add Docker GPG key"
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null || handle_error "Failed to set up Docker repository"
    sudo apt-get update || handle_error "Failed to update package list after adding Docker repository"
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io || handle_error "Failed to install Docker"
    sudo systemctl start docker || handle_error "Failed to start Docker service"
    sudo systemctl enable docker || handle_error "Failed to enable Docker service"
}

# Function to install Docker Compose
install_docker_compose() {
    echo "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose || handle_error "Failed to download Docker Compose"
    sudo chmod +x /usr/local/bin/docker-compose || handle_error "Failed to make Docker Compose executable"
}

# Check if Docker is installed
if ! command_exists docker; then
    install_docker
else
    echo "Docker is already installed."
fi

# Check if Docker Compose is installed
if ! command_exists docker-compose; then
    install_docker_compose
else
    echo "Docker Compose is already installed."
fi

# Clone Mythic repository
if [ ! -d "Mythic" ]; then
    echo "Cloning Mythic repository..."
    git clone https://github.com/its-a-feature/Mythic.git || handle_error "Failed to clone Mythic repository"
else
    echo "Mythic repository already cloned."
    echo "Do you want to update the Mythic repository? (y/n)"
    read -r update_choice
    if [ "$update_choice" == "y" ]; then
        cd Mythic || handle_error "Failed to change directory to Mythic"
        git pull || handle_error "Failed to update Mythic repository"
        cd ..
    fi
fi

# Navigate to Mythic directory
cd Mythic || handle_error "Failed to change directory to Mythic"

# Install dependencies using the provided script
echo "Installing Mythic dependencies..."
sudo ./install_docker_ubuntu.sh || handle_error "Failed to install Mythic dependencies"

# Copy and edit .env file
if [ ! -f ".env" ]; then
    echo "Copying .env file..."
    cp .env.example .env || handle_error "Failed to copy .env file"
    echo "Please edit the .env file as necessary."
else
    echo ".env file already exists."
fi

# Start Mythic
echo "Starting Mythic..."
sudo ./start_mythic.sh || handle_error "Failed to start Mythic"

# Inform the user
echo "Mythic C2 should now be running. Access it at https://localhost:7443 (default credentials are username: mythic_admin, password: mythic_password)"

# Function to update Mythic
update_mythic() {
    echo "Updating Mythic..."
    git pull || handle_error "Failed to pull latest Mythic changes"
    sudo ./stop_mythic.sh || handle_error "Failed to stop Mythic"
    sudo ./start_mythic.sh || handle_error "Failed to start Mythic"
}

# Prompt to update Mythic if it already exists
if [ -d "Mythic" ]; then
    echo "Do you want to update Mythic? (y/n)"
    read -r update_choice
    if [ "$update_choice" == "y" ]; then
        update_mythic
    else
        echo "Skipping Mythic update."
    fi
fi

echo "Mythic C2 installation and setup complete."
