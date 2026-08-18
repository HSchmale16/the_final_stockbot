FROM public.ecr.aws/lambda/python:3.12

# Install system dependencies including Tesseract OCR, developer tools, and graphics libraries
RUN yum update -y && \
    yum install -y \
    tesseract \
    tesseract-langpack-eng \
    ghostscript \
    mesa-libGL \
    && yum clean all

# Copy requirements and install python packages
COPY requirements.txt ${LAMBDA_TASK_ROOT}/
RUN pip install --no-cache-dir -r ${LAMBDA_TASK_ROOT}/requirements.txt

# Copy application source code
COPY app/ ${LAMBDA_TASK_ROOT}/app/

# Set the default CMD for the Lambda function handler
CMD ["app.handler.lambda_handler"]
