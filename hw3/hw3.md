## 1. Настройка ассихнронной репликации
  * Добавим +1 реплику в docker-compose 

    ![img.png](img.png)
  * На мастере создаем пользователя для репликации (выполнял на поднятой базе)

    ![img_1.png](img_1.png)
  * На реплике создаем конфиг мастера (на скрине должен был быть вариант с бинлогом, но я его стер ¯\_(ツ)_/¯ )

    ![img_2.png](img_2.png)
  * Включаем слейв

    ![img_3.png](img_3.png)
  * Тут должен был быть скрин с show slave status, но я не успел его сделать с бинлогом ¯\_(ツ)_/¯ 

## 2. Нагрузочный тест
  * Пустим нагрузку на мастер (вывод -- нагрузка пошла)

    ![img_4.png](img_4.png)
  * Пустим нагрузку на реплику (к сожалению ничего умнее поменять урл до базы я не придумал)
     Нагрузка на реплику

    ![img_5.jpg](img_5.jpg)
     Нагрузка на мастер

    ![img_6.jpg](img_6.jpg)
    Вывод -- нагрузка ушла в реплику

## 3. Настройка row-based GTID и 2 слейва
   * Добавляем +1 слейв (тривиально)
   * Переношу данные дампом на него (с помощью IDE)
   * В конфиге мастера вклчючаем row_based и GTID 

     ![img_5.png](img_5.png)
   * В конфигах реплик проделываем то же самое.
   * Стопим репликацию, меняем на auto_position в мастере

     ![img_6.png](img_6.png)
   * Смотрим, что теперь репликация в GTID режиме
     
     ![img_7.png](img_7.png)
   * Установка и включение semi-sync. Мастер, на репликах аналогично, но с master -> slave
    ![img_12.png](img_12.png)

## 3. Эксперимент с нагрузкой и потерей данных
   * Пустим нагрузку с регистрацией

   ![img_8.png](img_8.png)
   * Убиваю инстанс с docker kill (по идее то же самое, что просто kill), ожидаемо приложение не может записать

     ![img_9.png](img_9.png)
   * Я добавил счетчик, но по итогу просто посмотреть через count(*), сколько записалось
   * Master

        ![img_13.png](img_13.png)
   * Slave 1

        ![img_14.png](img_14.png)
   * Slave 2

        ![img_15.png](img_15.png)
   * Вывод -- в свежий слейв транзации не потерялись, во 2 слейв несколько транзакций потерялось.
   * Переключение свежего слейва на мастер делал так. 2 слейв переключил тривиально, как и выше, но с лругим урлом

        ![img_16.png](img_16.png)


   Так как я много эксперементировал с выполнением команд на поднятой базе, то просто приложу весь код из data-grip, который остался.