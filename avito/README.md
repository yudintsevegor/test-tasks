# April 2019

# Тестовое задание для BE стажера

Необходимо реализовать сервис на Golang, показывающий текущее время в разных временных зонах.

Сервис реализует JSON API работающее по HTTP. На вход принимает список зон, в ответе выдает список зон с текущим временем в них.

Задача **может** быть решена со следующими усложнениями:

1. Сервис поставляется как Docker образ, опубликованный в публичном репозитории
2. Реализация покрыта тестами
3. Запрос и ответ проходят JSON валидацию по JSON Schema
4. Сервис реализует сохранение набора зон для пользователя и их отображение по идентификатору пользователя. Хранилище может быть не персистентным.

Что дает решение задачи с усложнениями: + в карму равный номеру усложнения, т.е. +1 за 1-й пункт, +2 за второй и т.д.
