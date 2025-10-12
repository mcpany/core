# Use a slim Python image
FROM python:3.11-slim

# Set the working directory
WORKDIR /app

# Copy the application code
COPY main.py .

# Install dependencies
RUN pip install --no-cache-dir cowsay "fastmcp>=2.0.0"

# Command to run the application
CMD ["python3", "-u", "main.py", "--mcp-stdio"]