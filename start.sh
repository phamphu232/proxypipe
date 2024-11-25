#!/bin/sh

# Check if the file proxies.txt exists
if [ ! -f proxies.txt ]; then
  echo "File proxies.txt does not exist!"
  exit 1
fi

# Initialize bind port value starting from 8001
bind_port=8001

# Remove special characters ^M from the proxies file
sed -i 's/\r//g' proxies.txt

# Read each line in the proxies.txt file
while IFS= read -r proxy || [ -n "$proxy" ]; do
    if [[ "$proxy" =~ ^[^#].* ]]; then
    # Output the current proxy to check
    # echo "Using proxy: $proxy"

    # Extract the proxy type (http or socks5)
    proxy_type=$(echo "$proxy" | cut -d':' -f1)

    # Extract username and password from the proxy (before the @ symbol)
    user_password=$(echo "$proxy" | sed -e 's|^[a-zA-Z0-9]*://||' | cut -d'@' -f1)
    user=$(echo "$user_password" | cut -d':' -f1)
    password=$(echo "$user_password" | cut -d':' -f2)

    # Extract target (host:port) after the @ symbol
    target=$(echo "$proxy" | sed -e 's|^[a-zA0-9]*://.*@||' | cut -d'/' -f1)
    host=$(echo "$target" | cut -d':' -f1)
    port=$(echo "$target" | cut -d':' -f2)

    # # Output the extracted information for checking
    # echo "proxy_type: $proxy_type"
    # echo "user: $user"
    # echo "password: $password"
    # echo "target: $host"
    # echo "port: $port"
    # echo "bind port: $bind_port"

    # Call the proxypipe program with the extracted arguments and incrementing bind port
    ./proxypipe -t "$host:$port" -u "$user" -p "$password" -b "127.0.0.1:$bind_port" -type "$proxy_type" &

    # Increment bind port value for the next call
    bind_port=$((bind_port + 1))

    # Add a short wait time between calls to avoid overloading
    sleep 2
  fi
done < proxies.txt

# Keep the script running indefinitely
tail -f /dev/null
