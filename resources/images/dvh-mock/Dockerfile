# Use the official Nginx image as the base image
FROM nginxinc/nginx-unprivileged:stable-alpine-slim

# Copy the Nginx configuration file to the container
COPY resources/images/dvh-mock/nginx.conf /etc/nginx/conf.d/default.conf

# Copy the JSON file to the container
COPY resources/images/dvh-mock/dmo_ops_tnsnames.json /var/run/dmo_ops_tnsnames.json

# Expose port 8080
EXPOSE 8080

# Start Nginx when the container starts
CMD ["nginx", "-g", "daemon off;"]