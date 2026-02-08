# DonWeather-ms-georesolve

Микросервис геолокации для проекта DonWeather: определение ближайшего города по координатам (latitude/longitude) с помощью PostGIS. Clean Architecture: delivery, usecase, repository.

---

## Установка (локальная)

### Требования

- Go 1.24+
- Docker и Docker Compose
- PostgreSQL с расширением PostGIS (или через Docker Compose)

### 1. Клонирование и зависимости

```bash
git clone <repository-url>
cd DonWeather-ms-georesolve
go mod download
```

### 2. Запуск PostgreSQL с PostGIS

```bash
docker compose up -d
```

### 3. База данных

Подключитесь к PostgreSQL (локально или через `kubectl port-forward`), затем:

```sql
CREATE USER georesolve WITH PASSWORD 'mypassword';
CREATE DATABASE georesolve OWNER georesolve;
\c georesolve
CREATE EXTENSION IF NOT EXISTS postgis;
GRANT USAGE, CREATE ON SCHEMA public TO georesolve;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO georesolve;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO georesolve;
\q
```

Таблица `cities` с геометрией (PostGIS):

```sql
CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    geom geometry(Point, 4326) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cities_geom ON cities USING GIST (geom);
```

### 4. Переменные окружения

```bash
export DB_PASSWORD='mypassword'
export DB_USER='georesolve'
export DB_HOST='localhost'
export DB_PORT='5432'
export DB_NAME='georesolve'
```

### 5. Запуск

```bash
go run cmd/main.go
```

Сервис: `http://localhost:8080`.

### 6. API

**POST** `/api/v1/georesolve` — определение ближайшего города по координатам.

```bash
curl -X POST "http://localhost:8080/api/v1/georesolve" \
  -H "Content-Type: application/json" \
  -d '{"latitude": 47.2357, "longitude": 39.7015}'
```

**Health-check:**

```bash
curl http://localhost:8080/health/liveness
curl http://localhost:8080/health/readiness
```

---

## Развёртывание в Dev кластер

Развёртывание через Argo CD. Манифест Application применяется **после** создания секретов в Vault и базы данных.

### Порядок развёртывания

1. **Секреты в Vault** — создать секрет `secret/donweather-ms-georesolve` (см. подраздел «Секреты в Vault» ниже).
2. **База данных** — создать БД и пользователя (см. подраздел «База данных» ниже).
3. **Application** — применить манифест из репозитория:
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   kubectl apply -f .argocd/application.yaml
   ```
4. **Загрузка городов** — при первом `helm install` автоматически запускается Job (`pre-install` hook), который создаёт таблицу `cities` с PostGIS-индексом и загружает справочник ~1083 городов из `data/koord_russia.csv`. Если таблица уже содержит данные — импорт пропускается.

### Секреты в Vault

Путь в Vault (KV v2): **`secret/donweather-ms-georesolve`**. VaultStaticSecret создаётся Helm chart при установке (`vaultSecretOperator.enabled: true`).

**Ключи:** `db-password`, `db-user`, `db-host`, `db-port`, `db-name` (в поде: `DB_PASSWORD`, `DB_USER`, `DB_HOST`, `DB_PORT`, `DB_NAME`).

**Шаги:**

1. Переменные (Services кластер):
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   export VAULT_ADDR="http://127.0.0.1:8200"
   export VAULT_TOKEN=$(cat /tmp/vault-root-token.txt)
   ```

2. Включить KV v2 по пути `secret` (если нужно):
   ```bash
   kubectl exec -it vault-0 -n vault -- sh -c "
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='$VAULT_TOKEN'
   vault secrets enable -version=2 -path=secret kv 2>&1 || echo 'Уже включен'
   "
   ```

3. Создать секрет (подставить свои значения; пароли в одинарных кавычках):
   ```bash
   kubectl exec -it vault-0 -n vault -- sh -c "
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='$VAULT_TOKEN'
   vault kv put secret/donweather-ms-georesolve \
     db-password='<ПАРОЛЬ_БД>' \
     db-user='<USER_БД>' \
     db-host='<ХОСТ_POSTGRESQL>' \
     db-port='5432' \
     db-name='georesolve'
   "
   ```

4. Проверить: `vault kv get secret/donweather-ms-georesolve` (в том же `kubectl exec` с `VAULT_ADDR` и `VAULT_TOKEN`).

5. Обновление — снова `vault kv put secret/donweather-ms-georesolve ...`. После смены секрета при необходимости: `kubectl rollout restart deployment donweather-ms-georesolve -n donweather`.

### База данных

PostgreSQL для Dev может быть в Services кластере (доступ по LoadBalancer или port-forward). Пользователь, пароль и имя БД должны совпадать с теми, что записаны в секрете Vault (`db-user`, `db-password`, `db-name`).

1. Подключиться к PostgreSQL (подставьте хост/порт из секрета или из окружения кластера):
   ```bash
   # Например, через port-forward к сервису в Services кластере
   kubectl port-forward -n postgresql svc/postgresql 5432:5432
   psql -h localhost -U postgres -d postgres
   ```

2. Создать пользователя и базу (имя пользователя и пароль — как в ключах `db-user` и `db-password` в Vault; имя базы — как `db-name`, обычно `georesolve`):
   ```sql
   CREATE USER georesolve WITH PASSWORD 'ваш_пароль';
   CREATE DATABASE georesolve OWNER georesolve;
   \c georesolve
   CREATE EXTENSION IF NOT EXISTS postgis;
   GRANT USAGE, CREATE ON SCHEMA public TO georesolve;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO georesolve;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO georesolve;
   ```

3. Таблица `cities` и загрузка справочника городов создаются автоматически при первом деплое через Helm pre-install Job (см. «Порядок развёртывания»). Ручная инициализация не требуется.

### Application

Применить манифест (из корня клонированного репозитория):

```bash
kubectl apply -f .argocd/application.yaml
```

---

## Структура проекта

```
├── cmd/main.go                          — точка входа
├── internal/
│   ├── delivery/
│   │   ├── http/                        — HTTP-обработчики
│   │   └── middleware/                   — middleware (tracing)
│   ├── usecase/                         — бизнес-логика
│   ├── repository/                      — работа с БД (PostGIS)
│   └── domain/                          — доменные модели
├── .argocd/application.yaml             — манифест Argo CD Application
├── data/koord_russia.csv                — справочник городов РФ (CSV)
├── .helm/donweather-ms-georesolve/      — Helm chart (включая db-init Job)
├── Dockerfile                           — сборка Docker-образа
├── Jenkinsfile                          — CI-пайплайн
└── README.md
```

## Лицензия

**BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
