# Текущая структура проекта:
- Domain Models (domain/models/) - место для бизнес-сущностей.
- Services (domain/services/) - содержит бизнес-логику (это и есть use cases).
- Handlers (internal/http/handlers/) - это контроллеры.
- DTOs (internal/http/dto/) (internal/repository/dto/) - разделение на транспортных модели.
- Repository - слой для работы с данными.
