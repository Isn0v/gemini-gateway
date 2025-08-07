FROM python:3.9-slim

ENV CONNECTION_WAY=cloud
WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY client.py .

CMD ["python", "client.py"]