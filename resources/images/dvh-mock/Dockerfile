# Use the official Nginx image as the base image
FROM nginx:1.27.3

# Copy the Nginx configuration file to the container
COPY resources/images/dvh-mock/nginx.conf /etc/nginx/conf.d/default.conf

# Copy the JSON file to the container
COPY resources/images/dvh-mock/dmo_ops_tnsnames.json /home/nada/dmo_ops_tnsnames.json

# Expose port 80
EXPOSE 80

# Start Nginx when the container starts
CMD ["nginx", "-g", "daemon off;"]