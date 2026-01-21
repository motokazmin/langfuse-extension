import os
import requests
import json
import uuid
from datetime import datetime, timezone
from dotenv import load_dotenv
import sys

print("Загрузка переменных окружения из .env файла...")
load_dotenv()

LANGFUSE_PUBLIC_KEY = os.getenv("LANGFUSE_PUBLIC_KEY")
LANGFUSE_SECRET_KEY = os.getenv("LANGFUSE_SECRET_KEY")
LANGFUSE_HOST = os.getenv("LANGFUSE_BASEURL", "https://cloud.langfuse.com")

if not all([LANGFUSE_PUBLIC_KEY, LANGFUSE_SECRET_KEY]):
    print("❌ Ошибка: Не все переменные окружения установлены.")
    sys.exit(1)

print("Переменные окружения успешно загружены.")

def create_trace(name, user_id, metadata=None):
    """Создаёт трейс через HTTP API"""
    url = f"{LANGFUSE_HOST}/api/public/ingestion"
    
    trace_id = str(uuid.uuid4())
    timestamp = datetime.now(timezone.utc).isoformat()
    
    payload = {
        "batch": [{
            "id": str(uuid.uuid4()),
            "type": "trace-create",
            "timestamp": timestamp,
            "body": {
                "id": trace_id,
                "name": name,
                "userId": user_id,
                "metadata": metadata or {},
                "timestamp": timestamp
            }
        }]
    }
    
    response = requests.post(
        url,
        auth=(LANGFUSE_PUBLIC_KEY, LANGFUSE_SECRET_KEY),
        headers={"Content-Type": "application/json"},
        json=payload
    )
    
    if response.status_code == 207:
        print(f"✅ Трейс '{name}' создан (ID: {trace_id})")
        return trace_id
    else:
        print(f"❌ Ошибка создания трейса: {response.status_code} - {response.text}")
        return None

def create_span(trace_id, name, input_data, output_data, level="DEFAULT", latency_ms=100):
    """Создаёт span в трейсе"""
    url = f"{LANGFUSE_HOST}/api/public/ingestion"
    
    span_id = str(uuid.uuid4())
    start_time = datetime.now(timezone.utc)
    end_time = datetime.now(timezone.utc)
    
    payload = {
        "batch": [{
            "id": str(uuid.uuid4()),
            "type": "span-create",
            "timestamp": start_time.isoformat(),
            "body": {
                "id": span_id,
                "traceId": trace_id,
                "name": name,
                "startTime": start_time.isoformat(),
                "endTime": end_time.isoformat(),
                "input": input_data,
                "output": output_data,
                "level": level
            }
        }]
    }
    
    response = requests.post(
        url,
        auth=(LANGFUSE_PUBLIC_KEY, LANGFUSE_SECRET_KEY),
        headers={"Content-Type": "application/json"},
        json=payload
    )
    
    if response.status_code == 207:
        print(f"  ✅ Span '{name}' добавлен")
        return span_id
    else:
        print(f"  ❌ Ошибка создания span: {response.status_code}")
        return None

def create_generation(trace_id, name, model, input_tokens, output_tokens, cost):
    """Создаёт generation в трейсе"""
    url = f"{LANGFUSE_HOST}/api/public/ingestion"
    
    gen_id = str(uuid.uuid4())
    timestamp = datetime.now(timezone.utc).isoformat()
    
    payload = {
        "batch": [{
            "id": str(uuid.uuid4()),
            "type": "generation-create",
            "timestamp": timestamp,
            "body": {
                "id": gen_id,
                "traceId": trace_id,
                "name": name,
                "model": model,
                "startTime": timestamp,
                "endTime": timestamp,
                "usage": {
                    "input": input_tokens,
                    "output": output_tokens
                },
                "metadata": {
                    "calculatedCost": cost
                }
            }
        }]
    }
    
    response = requests.post(
        url,
        auth=(LANGFUSE_PUBLIC_KEY, LANGFUSE_SECRET_KEY),
        headers={"Content-Type": "application/json"},
        json=payload
    )
    
    if response.status_code == 207:
        print(f"  ✅ Generation '{name}' добавлен")
        return gen_id
    else:
        print(f"  ❌ Ошибка создания generation: {response.status_code}")
        return None

print("\nНачинаем генерацию тестовых трейсов...\n")

# 1. Трейс с ошибкой
print("1/4: Создаём трейс с ошибкой...")
trace_id = create_trace("failed-tool-call", "user-error", {"tags": ["test-data", "error"]})
if trace_id:
    create_span(
        trace_id, 
        "call-external-api",
        {"url": "http://non-existent-service.local"},
        {"error": "Failed to connect to host"},
        level="ERROR"
    )

# 2. Трейс с высокой задержкой
print("\n2/4: Создаём трейс с высокой задержкой...")
trace_id = create_trace("performance-bottleneck", "user-latency", {"tags": ["test-data", "performance"]})
if trace_id:
    create_span(
        trace_id,
        "fast-step-1",
        {},
        {},
        latency_ms=100
    )
    create_span(
        trace_id,
        "slow-database-query",
        {"query": "SELECT * FROM huge_table"},
        {"rows": 1000000},
        latency_ms=6000
    )

# 3. Трейс с высокой стоимостью
print("\n3/4: Создаём трейс с высокой стоимостью...")
trace_id = create_trace("high-cost-report", "user-cost", {"tags": ["test-data", "cost"]})
if trace_id:
    create_generation(
        trace_id,
        "expensive-summary",
        "gpt-4-turbo",
        20000,
        5000,
        0.25
    )

# 4. Трейс с циклом
print("\n4/4: Создаём трейс с логическим циклом...")
trace_id = create_trace("logical-loop-agent", "user-loop", {"tags": ["test-data", "loop"]})
if trace_id:
    for i in range(5):
        create_span(
            trace_id,
            "search-tool",
            {"query": "what is langfuse"},
            {"result": "Langfuse is an open-source observability platform..."},
            latency_ms=100
        )

print("\n🎉 Все тестовые данные отправлены в Langfuse!")
print("Обнови страницу в браузере, чтобы увидеть их.")