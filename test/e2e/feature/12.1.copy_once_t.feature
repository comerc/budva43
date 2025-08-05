Feature: 12.1.copy_once_t

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение не обновляется в целевом чате (версионирование)
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "some text"
    And будет текст "some text"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    # And сообщение в чате "DST PRV GRP 1" (4897079215)
    # ^ 400 Message links are available only for messages in supergroups and channel chats
    Given сброс проверок
    And сообщение с текстом "some OTHER text"
    And будет ссылка на предыдущую версию сообщения в целевом чате
    And будет предыдущее сообщение со ссылкой на сообщение в целевом чате
    When пользователь редактирует сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    # And сообщение в чате "DST PRV GRP 1" (4897079215)
    # ^ 400 Message links are available only for messages in supergroups and channel chats

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |
# | SRC PRV GRP 1 | 4832061506    |
# ^ 400 Message links are available only for messages in supergroups and channel chats

