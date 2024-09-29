# Paygo

Сервис для приема платежей через Yoomoney

## Установка

Запускаем в докере

```sh
docker pull registry.gitflic.ru/project/kovardin/payments/payments
docker run -v /home/user/data:/data  -p 8080:8080 registry.gitflic.ru/project/kovardin/payments/payments:latest --dir /data --dev  --http :8080 serve 
```

`/home/user/data` - дирректория, где будут находиться все данные сервиса

## Подключение SDK

Android SDK доступно тут [depot.kovardin.ru](https://depot.kovardin.ru/)

Подключаем репозиторий:

```gradle
repositories {
    maven {
        name = "depot"
        url = "https://depot.kovardin.ru/packages"
    }
}
```

Подключаем зависимость:

```gradle
dependencies {
    implementation 'ru.kovardin:billing:0.1.3'
}
```