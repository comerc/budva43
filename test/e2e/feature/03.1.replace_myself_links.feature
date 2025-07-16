@who:user
@what:links
@why:cleanliness

Feature: 03.1.replace_myself_links

  Background:
    Given будет пересылка - копия

  Scenario Outline: Ссылки на свои сообщения заменяются
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    When пользователь отправляет сообщение
    Then пауза 10 сек.
    And сообщение в чате
    And сообщение в чате "-1002667730628" (DST PUB CHL 1)
    And сообщение в чате "-1002866470933" (DST PUB GRP 1)
    Given сообщение со ссылкой на последнее сообщение
    And будет замена ссылки на сообщение в целевом чате
    When пользователь отправляет сообщение
    Then пауза 10 сек.
    And сообщение в чате "-1002667730628" (DST PUB CHL 1)
    And сообщение в чате "-1002866470933" (DST PUB GRP 1)

    Examples:
      | src_chat_id    | src_chat_name |
      | -1002641439846 | SRC PUB CHL 1 |
      | -1002792282007 | SRC PRV CHL 1 |
      | -1002736661856 | SRC PUB GRP 1 |
