# yandexCalculator
Распределенный вычислитель арифметических выражений
Разработал - Антонов Денис (телеграм для связи и вопросов @semyadmin)

<details>
    <summary>Описание задачи: </summary>
    Пользователь хочет считать арифметические выражения. Он вводит строку 2 + 2 * 2 и хочет получить в ответ 6. Но наши операции сложения и умножения (также деления и вычитания) выполняются "очень-очень" долго. Поэтому вариант, при котором пользователь делает http-запрос и получает в качетсве ответа результат, невозможна. Более того: вычисление каждой такой операции в нашей "альтернативной реальности" занимает "гигантские" вычислительные мощности. Соответственно, каждое действие мы должны уметь выполнять отдельно и масштабировать эту систему можем добавлением вычислительных мощностей в нашу систему в виде новых "машин". Поэтому пользователь, присылая выражение, получает в ответ идентификатор выражения и может с какой-то периодичностью уточнять у сервера "не посчиталость ли выражение"? Если выражение наконец будет вычислено - то он получит результат. Помните, что некоторые части арфиметического выражения можно вычислять параллельно.

    Front-end часть

    GUI, который можно представить как 4 страницы

        Форма ввода арифметического выражения. Пользователь вводит арифметическое выражение и отправляет POST http-запрос с этим выражением на back-end. Примечание: Запросы должны быть идемпотентными. К запросам добавляется уникальный идентификатор. Если пользователь отправляет запрос с идентификатором, который уже отправлялся и был принят к обработке - ответ 200. Возможные варианты ответа:
            200. Выражение успешно принято, распаршено и принято к обработке
            400. Выражение невалидно
            500. Что-то не так на back-end. В качестве ответа нужно возвращать id принятного к выполнению выражения.
        Страница со списком выражений в виде списка с выражениями. Каждая запись на странице содержит статус, выражение, дату его создания и дату заверщения вычисления. Страница получает данные GET http-запрсом с back-end-а
        Страница со списком операций в виде пар: имя операции + время его выполнения (доступное для редактирования поле). Как уже оговаривалось в условии задачи, наши операции выполняются "как будто бы очень долго". Страница получает данные GET http-запрсом с back-end-а. Пользователь может настроить время выполения операции и сохранить изменения.
        Страница со списком вычислительных можностей. Страница получает данные GET http-запросом с сервера в виде пар: имя вычислительного ресурса + выполняемая на нём операция.

        Требования:
        Оркестратор может перезапускаться без потери состояния. Все выражения храним в СУБД.
        Оркестратор должен отслеживать задачи, которые выполняются слишком долго (вычислитель тоже может уйти со связи) и делать их повторно доступными для вычислений.


    Back-end часть

    Состоит из 2 элементов:

        Сервер, который принимает арифметическое выражение, переводит его в набор последовательных задач и обеспечивает порядок их выполнения. Далее будем называть его оркестратором.
        Вычислитель, который может получить от оркестратора задачу, выполнить его и вернуть серверу результат. Далее будем называть его агентом.

    Оркестратор
    Сервер, который имеет следующие endpoint-ы:

        Добавление вычисления арифметического выражения.
        Получение списка выражений со статусами.
        Получение значения выражения по его идентификатору.
        Получение списка доступных операций со временем их выполения.
        Получение задачи для выполения.
        Приём результата обработки данных.


    Агент
    Демон, который получает выражение для вычисления с сервера, вычисляет его и отправляет на сервер результат выражения. При старте демон запускает несколько горутин, каждая из которых выступает в роли независимого вычислителя. Количество горутин регулируется переменной среды.
</details>

Запуск оркестратора и агента (агентов) можно выполнить несколькими способами
Во всех способах надо сначала перейти на страницу https://github.com/adminsemy/yandexCalculator

1. Простой способ с использование докер файлов.

Нужно, что бы у вас был установлен docker. Можно поставить его по инструкциям

https://docs.docker.com/engine/install/

Делаем клон репозитория https://github.com/adminsemy/yandexCalculator

Если установлен git

    git clone https://github.com/adminsemy/yandexCalculator

Если git не установлен
Качаем архив по ссылке 

    https://github.com/adminsemy/yandexCalculator/archive/refs/heads/main.zip

и раcпаковываем в любую папку.
Далее переходим в коревую папку и запускаем команду. (ВНИМАНИЕ - порты 8080, 7777 и 5433 должны быть не заняты!
Если они заняты и нужны, то надо зайти в папку Orchestrator/config и в файле .env изменить на нужные а так же в файле docker-compose.yaml. Если меняется ORCHESTRATOR_TCP_PORT, то соответствующую настройку необходимо сделать в агенте (из корневой папки папка Agent/config и отредактировать параметр PORT))

    docker compose up

или 

    docker-compose up

После запуска контейнеров можно перейти по адресу

http://localhost:8080/

2. Второй способ без установки докера.

Так же делаем клон репозитория или качаем архив. Затем переходим в папку Orchestrator.В этой папке надо поменять настройки для запуска оркестратора: 
    В папке config меняем в файле .env ORCHESTRATOR_DB на 127.0.0.1 (или на вашу базу данных). База данных используется PostgreSQL
    Потом запускаем проект  командой
    go run cmd/orchestrator/main.go
    Возвращаемся в корневую папку
Затем нужны сделать настройки в агенте - из корневой папки проекта надо перейти в папку Agent/config и поменять настройку HOST=127.0.0.1
Затем из папки Agent запустить агент командой

    go run cmd/orchestrator/main.go

Запуск базы данных будет возложен на вас. Для корректной работы базы надо запустить в запущенной PostgreSQL файл yandex.sql из папки sql коревого каталога проекта. База данных должна называться orchestrator. Или вы можете поменять имя в настройках Orchestrator на любое другое.

3. Третий способ вообще без установленных программ.

Перейти в раздел Releases на github и скачать оттуда дистрибутивы для Linux, MacOS и Windows. Дистрибутивы для x64 систем.
Или скачать сразу по ссылке

    https://github.com/adminsemy/yandexCalculator/files/14321253/Distrib.zip

Распаковать и запустить нужный файл для оркестратора и агента из директории вашей операционной системы (для lInux и Windows файлы точно работают). Можно так же произвести необходимые настройки для запуска приложения (настройки в папках config)

#Настройки оркестратора

Все настройки находятся в файле ./Orchestrator/config/.env и для докер образа в файле docker-compose.yaml.
ORCHESTRATOR_HTTP_PORT - настройка HTTP порта для работы сервера по протоколу http.
ORCHESTRATOR_TCP_PORT - настройка для работы TCP сервера. TCP сервер нужен для работы с агентом.
ORCHESTRATOR_DB - адрес базы данных.
ORCHESTRATOR_DB_NAME - название базы данных.
ORCHESTRATOR_DB_PORT - порт для подключения к базе данных.
ORCHESTRATOR_DB_USER - имя пользователя для базы данных.
ORCHESTRATOR_DB__PASSWORD - пароль для базы данных.

#Настройки агента

Все настройки находятся в файле ./Agent/config/.env и для докер образа в файле docker-compose.yaml.
HOST - имя или IP адрес для оркестратора.
PORT - порт оркестратора.
MAX_GOROUTINES_AGENT - количество горутин у агента.

#Описание API оркестратора

Можно использовать curl, Postman или что-нибудь еще для отправки данных да оркестратор. Так же используется визуальное отображение через веб-браузер.
Используются следующие API:

"/" - любой запрос вернет html файл сгенерированных для подключения к оркестратору.

"/duration" - метол POST устанавливает нужное время для каждой арифметической операции. Метод GET вернет текущие настройки времени. 
Если настроек нет - вернет значения по умолчанию.

"/expression" - метод POST принимает нужное выражение. Поддерживаются операции "+", "-", "*", "/", а так же скобки "(" и ")". Все пробелы игнорируются (удаляются). На данный момент не поддерживаются унарные операции (но в процессе вычисления могут быть данные ниже 0) - можно использовать 0-10 (например), возведение в степень и т.д. Если значение некорректное - система вернет статус 400 и описание ошибки. Если все верное - начнется высчитываться выражение и вернет выражение с его ID, датой его создания, датой завершения и его текущий статус. Если написать выражение, которое уже было, то вернется это выражение, его ID и статус

"/id/{id}" - принимает ID выражения и возвращает его, если оно найдено. Если нет - возвращается статс 400 и ошибка.

"/workers" - по методу GET возвращает количество запущенных агентов, общее количество горутин всех агентов, занятых горутин (ворекров), которые обрабатывают текущую операцию и массив обрабатываемых на данный момент операций.

Основное меню программы через веб браузер (http://localhost:8080/ по умолчанию)


![Выражение](/MDImages/01.png)
![Список выражений](/MDImages/02.png)

Меню остальных вкладок интуитивно понятно

#Принцип работы программы

Подсчет выражения.
Пользователь вводит выражение. Оркестратор принимает выражение,валидирует его и, если валидация проходит, создает AST дерево, разбивая выражение на операции. Затем выражение сохраняется в мапе (если есть соединения с базой данных - выражение сохраняется в базу) и запускается вычисление выражения. Каждая отдельная операция ставится в очередь по порядку добавления и ждет, когда агент попросит данных. Когда воркер агента подключается к оркестратору (каждый воркер подключается в своей горутине), запрашивая данные, данные извлекаются из очереди по порядку, записываются в мапу очереди и отправляются агенту для вычисления. Воркеры агента не сохраняют соединение с орекстатором. После вычисления, агнет возвращает результат операции или ошибку, если операцию не удалось посчитать. Оркестратор принимает операцию, сообщает выражению результат вычисления и сохраняет в мапу вычисленных операций ее состояние. Так же в отдельную мапу сохраняется выражение с подсчитанной операцией для сохранения в базу данных (сохраняется каждую секунду). Дальше все повторяется до тех пор, пока выражение не будет посчитано полностью. Полностью подсчитанное выражение остается в мапе сохраненных выражений.

Установка времени подсчета выражений.
Пользователь присылает данных - оркестратор проверяет эти данные и если они верные - перезаписывает и пытается сохранить в базу данных.

Мониторинг агентов.
Каждый агент соединяется с оркестратором и держит постоянною связь, периодически сообщая о том, сколько воркеров на данный момент работает и сколько всего есть. Ворекры так же соединяются с оркестратором для запроса операций, но связь не сохраняют. Когда агент отключается - вся информация о работающих на нем процессах и их количество обнуляется.
