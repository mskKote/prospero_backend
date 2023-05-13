# prospero_back-end

---

## Запуск
```shell
# graylog & grafana
docker-compose up

# перезапуск prospero
docker-compose up -d --no-deps --build prospero
```

## Инфраструктура

При локальном запуске

[Kibana](http://127.0.0.1:5601/) | 
[Prometheus](http://localhost:9090/) | 
[Grafana](http://localhost:3000/) | 
[Jaeger](http://localhost:16686/)

```shell
# example Graylog
echo -n '{ "version": "1.1", "host": "example.org", "short_message": "TEST #2", "level": 5, "_some_info": "foo
" }' | nc -w0 -u localhost 12201
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
