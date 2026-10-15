# Изолированные E2E-сценарии

Эти сценарии не используют рабочую, demo или production-базу. Перед запуском
нужно поднять отдельный Compose-проект и получить переменные окружения:

```bash
./scripts/e2e/prepare-isolated.sh
# скрипт выводит export-команды; выполните их в том же shell
cd frontend && npm run test:e2e:smoke
```

`prepare-isolated.sh` создаёт отдельные Docker-тома и вызывает `./seed` только
в контейнере `mentra-e2e`. Seed намеренно очищает данные, поэтому скрипт
отказывается использовать любое другое имя Compose-проекта. Он не запускает
Playwright.

Скрипты принимают только Docker context с локальным `unix://` или `npipe://`
endpoint и единственное имя проекта `mentra-e2e`. `DOCKER_HOST=ssh://…`, TCP
endpoint, а также другое имя проекта отклоняются до вызова Docker Compose.
`cleanup-isolated.sh` сам создаёт краткоживущий test-only env-файл, поэтому
корректно работает из новой shell без сохранённых E2E-переменных.

После проверки удалите только этот временный стек:

```bash
./scripts/e2e/cleanup-isolated.sh
```

## Наборы сценариев

| Набор | Команда | Данные | Изменения |
| --- | --- | --- | --- |
| Smoke | `npm run test:e2e:smoke` | Только изолированные | Нет |
| Extended | `E2E_RUN_EXTENDED=true npm run test:e2e:extended` | Только изолированные | Нет |
| Write | `E2E_ALLOW_WRITES=true npm run test:e2e:write` | Только изолированные | Да, только test stack |

Write-набор дополнительно проверяет `E2E_ISOLATED=true` и принимает только
локальный Docker stack с runtime-доказательством: Compose-проект и labels
контейнеров должны быть `mentra-e2e`, опубликованные порты должны совпадать с
`E2E_BASE_URL` и `E2E_API_URL`, а metadata фактически запущенного frontend
image должна содержать API URL, совпадающий с фактическим API mapping. Старый
image без metadata или с другим встроенным endpoint блокирует write-run. Это
предохраняет от случайной записи в staging, production или другой
локальный стек. Skipped-тесты не являются успешной
проверкой: для требуемого набора переменные должны быть заданы, иначе запуск
завершается ошибкой конфигурации.

Тестовый язык задаётся явно через `E2E_LOCALE` (по умолчанию `en-US`), часовой
пояс — через `E2E_TIMEZONE` (по умолчанию `Asia/Almaty`). Не используйте
локальное состояние браузера как источник данных для теста.

## Матрица следующего прогона

| Сценарий | Набор | Детерминированные данные | Условие готовности |
| --- | --- | --- | --- |
| Вход четырёх ролей | Smoke | `nu` + e2e super-admin | Готово к запуску |
| Health и форма входа | Smoke | Нет изменений данных | Готово к запуску |
| Просмотр курса студентом и преподавателем | Extended | Курсы и модули из seed | Готово к запуску |
| Создание модуля / правка урока | Write | Только `mentra-e2e` | Готово к запуску |
| Изоляция организаций | Extended | `nu` и `kbtu` из seed | После согласования API прав агента 1 |
| Приватные материалы и файлы | Extended | Upload volume `e2e_uploads` | После интеграции приватного file API |
| Экзамен: старт, восстановление, лимит попыток | Write | Экзамены seed | После финализации exam persistence |
| Публикация оценки и скрытый итог | Write + Extended | Задания seed | После финализации grade API |
| Уведомление после reload и dismissal | Extended | Стабильные notification events | После серверной persistence уведомлений |
| Logout/login другим пользователем | Extended | Две учётные записи `nu` | После security-интеграции |
| Google OAuth с организацией | Manual/stub | OAuth test client | Не запускать с реальным OAuth provider в CI |

Статус «готово к запуску» означает, что подготовлены данные и границы
окружения, но не означает, что браузерная проверка уже была выполнена.

## Учётные данные

Скрипт использует только тестовые учётные данные, заданные в `seed`:

- `E2E_STUDENT_EMAIL`, `E2E_STUDENT_PASSWORD`;
- `E2E_TEACHER_EMAIL`, `E2E_TEACHER_PASSWORD`;
- `E2E_ADMIN_EMAIL`, `E2E_ADMIN_PASSWORD`;
- `E2E_SUPER_ADMIN_EMAIL`, `E2E_SUPER_ADMIN_PASSWORD`.

Не добавляйте настоящие пароли в файлы репозитория или CI-логи.

## Граница AI-гарантии

В [docker-compose.e2e.yml](../../../docker-compose.e2e.yml) тестовый backend
жёстко использует `AI_PROVIDER=ollama` c адресом `http://ollama:11434`; ключ
OpenAI и внешний AI endpoint не передаются. Подготовленные Playwright-наборы
не вызывают AI-маршруты. Это исключает расход внешнего AI provider во время
обычных E2E-сценариев, но не делает Docker-стек полностью offline: Docker может
скачать образы, npm — зависимости при build. Не добавляйте AI-сценарии без
явного stub/provider-контракта.
