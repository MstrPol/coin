# semver-tag

Упрощённая tag-based модель для библиотек (`go-lib`, `java-maven-app`).

- **Trunk:** `main`.
- **Версия:** semver-теги `vMAJOR.MINOR.PATCH` без RC/snapshot qualifiers.
- **Publish:** только при semver-теге на коммите (`publish.when: tag`).

Для сервисов с release-ветками и RC используйте `trunk-based`.
