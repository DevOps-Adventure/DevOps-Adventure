# Use an official Python runtime as a parent image
FROM python:3.8-slim

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# If your requirements.txt is in a venv directory within your project, adjust the path accordingly
COPY requirements.txt /app/requirements.txt

# Install any needed packages specified in requirements.txt
RUN python -m pip install --upgrade pip
RUN pip install --no-cache-dir -r requirements.txt

# Make port 5000 available to the world outside this container
EXPOSE 5001

# Define environment variable
ENV NAME World

# Command to run the application
CMD ["python", "minitwit.py"]