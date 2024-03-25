# prospero_back-end

---

[![Prospero deployed](https://github.com/mskKote/prospero_backend/actions/workflows/deploy-job.yml/badge.svg)](https://github.com/mskKote/prospero_backend/actions/workflows/deploy-job.yml)

## Запуск
```shell
# graylog & grafana
docker compose up

# перезапуск prospero
docker compose up -d --no-deps --build prospero

# разработка
docker compose watch
```

## Инфраструктура

При локальном запуске

[Kibana](http://127.0.0.1:5601/) | 
[Prometheus](http://localhost:9090/) | 
[Grafana](http://localhost:3000/) | 
[Jaeger](http://localhost:16686/) |
[Swagger](http://localhost:80/swagger/index.html)

[//]: # (```shell)

[//]: # (# example Graylog)

[//]: # (echo -n '{ "version": "1.1", "host": "example.org", "short_message": "TEST #2", "level": 5, "_some_info": "foo)

[//]: # (" }' | nc -w0 -u localhost 12201)

[//]: # (```)

```shell
swag fmt && swag init -g ./cmd/main.go -o ./docs
```

---
## Архитектура

Следовал "чистой архитектуре" [по примеру](https://github.com/theartofdevel/golang-clean-architecture)

### internal

* controller
  * http/v1 - протокол/версионирование
* adapters
  * работа с базами (/bd)
  * кафкой (/kafka)
  * регистрация метрики (/metrics)
* domain
  * usecase - сборка нескольких сервисов
  * service - бизнес-логика для 1 сущности
  * entity - бизнес-сущность

### pkg

Общее между микросервисами. Клиенты баз, логгер, графана

* config - конфигурация сервиса
* logging - логгер
* metrics - middleware для gin

## Инструменты

* zap - логгер
* gin - роутинг
* [gofeed](https://github.com/mmcdole/gofeed) - парсер RSS
* [gocron](https://github.com/go-co-op/gocron) - запуск раз в N времени 
* [Планирование cron job](https://crontab.guru/#0_*_*_*_*) - правильно указать "N" для gocron

---
## Установка на машине

1. [docker](https://docs.docker.com/engine/install/ubuntu/)

2. [docker compose plugin](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-compose-on-ubuntu-22-04)

3. [SSH github](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account)

4. [GitHub actions runner](https://habr.com/ru/articles/737148/)