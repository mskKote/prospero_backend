# prospero_back-end


---

## Запуск
```shell
# graylog
docker-compose up
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
* domain
  * usecase - сборка нескольких сервисов
  * service - бизнес-логика для 1 сущности
  * entity - бизнес-сущность

### pkg

Общее между микросервисами. Клиенты баз, логгер, графана

* config - конфигурация сервиса
* logging - логгер
