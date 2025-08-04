Feature: 03.1.replace_myself_links

  Background:
    Given будет пересылка - копия

  Scenario Outline: Ссылки на свои сообщения заменяются
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    Given сообщение со ссылкой на последнее сообщение
    And будет ссылка на предыдущее сообщение в целевом чате
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |
# | SRC PRV GRP 1 | 4832061506    |
# ^ 400 Message links are available only for messages in supergroups and channel chats
