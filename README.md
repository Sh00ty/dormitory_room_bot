# Dormitory room bot
Телеграм бот созданный для решения насущных проблем, которые могут возникнуть в любой дружной компании.

Cсылка на работающего бота в телеграм - **https://t.me/domestosTbot**

- - -
## Как запускать?
Чтобы запустить своего бота необходимо выполнить следующие шаги:
- 1) Зайти в в telegram, создать в @BotFather нового бота и сохранить у себя его токен.
- 2) Поставить следующие переменные окружения в процессе в котором далее будет запускаться бот:
    ```
    # токен который вы получили на 1 шаге
    export tg_token=tg_token_from_botFather
    # переменная окружения, существует еще stg для тестирования. У них разные конфиги.
    export env=prod
    # пароль для postgres
    export pg_pass=your_secret_password1
    # пароль для mongo
    export mongo_pass=your_secret_password2
    ```
- 3) Если есть желание, то зайти в ```./internal/config/config.yaml``` и настроить все необходимые настройки для prod или stg окружения. По умолчанию там стоят валидные значения. Так же нужно их аккуратно сопоставить переменным из docker-compose.yaml.
- 4) Проверить и настроить по желанию ```docker-compose.yaml```.
- 5) Выполнить ```docker-composer up -d``` и затем ```docker logs -f dormitory_room_bot```. Если в течении минуты не появляется надпись app started, то попробуйте откомментировать ```#network_mode: host``` в docker-compose.yaml. Если ошибка не ушла, то я криво объяснил или были допущены ошибки при смене переменных в ```config.yaml``` или ```docker-compose.yaml```.
- - -
## Управление долгами
С помощью данного бота больше не придется в попыхах отправлять деньги друг другу после того как кто-то за всех заплатил.
Бот способен на протяжении долгих месяцев запоминать долги внутри одного чата и потом выдавать ***не более n-1 перевода*** для погашения абсолятно всех долгов за этот период. (где n - количество участников в группе)

## Обязанности и задачи
Так же с помощью бота можно создавать напоминалки для каких-то бытовых задач, например выкинуть мусор, задать пул исполнителей и поставить количество исполнителей (2 из 3 например) и затем менять этих двух человек чередовать людьми из этого пула. Так же есть задачи на которые можно подписываться/отписываться если интересуют ее нотификации.

## Общие списки
Если вы вдруг захотели с кем-то вести например спискок фильмов на посмотреть или просто списки с любой полезной информацией так чтобы у каждого это было под рукой и каждый мог добавить что-то туда или посмотреть то вам к dormitory room bot. А самое главное что можно выбрать несколько случайных элементов из любого списка.

## Краткое описание технических деталей
Стэк технологий:
- ***Go***
- ***PostgreSql***
- ***Prometheus***
- ***Grafana***
- ***Mongo-db***
- ***Docker***
- ***Gitlab-Ci***
- ***pgx***
- ***bash***
- ***docker-compose***

Некоторые интересные или нет подробности:
- Worker pool для параллельной обработки запросов
- Переотправления сообщений с прогрессивным таймаутом на случай каких-то сетевых сбоев итп 
- Транзакции в postgresql
- Pool соеденений с базой данных


# Список полезных команд:
- ```/register``` - регистрирует тебя в канале в котором ты напишешь это сообщение(необходимо для подсчета долгов)
- ```/bank``` - выводит баланс каждого участиника чата. Елси баланс меньше нуля значит что ты кому-то дожен, если больше то наоборот
- ```/credit @gera @steps @etc 3000``` - говорит, что исполнители после ```credit``` должны тому кто пишет 3000 деленное на их количество. Так же вместо 3000 можно написать простое математическое выражение в фигурных скобках, например ```/credit @gera @steps @etc {3000 * (299+1) / 2}```
- ```/cancel``` - отменяет долг созданный сообщением с ```/credit```. Для того, чтобы выбрать отменяемый долг, необходимо ответить на это сообщение с данной командой.
- ```/checkout``` - показывает кто кому сколько должен перевести внутри данного чата. После этого обнуляет все имеющиеся долги в этом чате
- ```/lists``` - выводит все имеющиеся списки
- ```/createl list_id item1 item2```... - создает список в котором уже буду лежать указанные элементы
- ```/addit list_id item``` - добавляет элемент в список
- ```/delit list_id 5``` - удаляет из списка элемент с заданным номером
- ```/randit list_id ```(количество опционально)- выводит случайный элемент из списка"
- ```/tasks``` - выводит все имеющиеся в этом чате задачи в виде нажимающихся кнопок
- ```/task имя_задачи @исполнитель1 @исп2 @исп3 2 2h 04-11-2022+21:08 тут большое описание``` - создает задачку где все опции указанные ниже не обязательные, их можно    писать, а можно и нет(но порядок в котором идут опции оч важен)

    @исполнитель1 ... это возможные исполнители данной задачи

    следующее число показывает сколько человек в моменте исполняет задачу
	исполнителей можно поменять или прокрутить по циклу если их количество равно цифре после них(лучше попробовать чтобы понять)

    далее идет интервал напоминания этой задачи, можно использовать 40s, 20m, 24h

    04-11-2022+21:08 - точные дата и время в которое бот напомнит о задаче
	после срабатывания точного времени начнет работать интервал

    и наконец сколь угодно большое описание
	
- ```/moment some_task @w1 @w2 @w3 2 2h 04-11-2022+21:08 большое описание``` - cоздает задачку которая исчезает после напоминания
- ```/subt some_task @w1 @w2 @w3 2 2h 04-11-2022+21:08 большое описание``` - cоздает задачку на которую можно подписываться и отписываться