import requests
import os

connection_way = os.getenv("CONNECTION_WAY", "local").lower()

SERVER_URL = f"http://localhost:8080/gemini"
if connection_way == "cloud":
  SERVER_URL = "http://gemini-gateway.local/gemini"

def ask_gemini(prompt):
	"""Отправляет запрос на сервер и получает ответ от Gemini."""
	try:
		response = requests.post(SERVER_URL, json={"prompt": prompt})
		response.raise_for_status() # Проверка на ошибки HTTP
		print("Ответ от Gemini:", response.json().get("response"))
	except requests.exceptions.RequestException as e:
		print(f"Ошибка при подключении к серверу: {e}")

if __name__ == "__main__":
	while True:
		user_prompt = input("Введите ваш запрос для Gemini (для выхода введите 'exit'): ")
		if user_prompt.lower() == 'exit':
			break
		ask_gemini(user_prompt)