@who:user
@what:links
@why:cleanliness

Feature: 03.2.delete_external_links

  Background:
    Given будет пересылка - копия

  Scenario Outline: Ссылки на внешние сообщения удаляются
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет YETI_MESSAGE
    Then ожидание 10 сек.
    And YETI_MESSAGE в чате
    # нет пересылки для YETI_MESSAGE
    Given сообщение со ссылкой на последнее сообщение
    And будет замена: ссылка на YETI_MESSAGE -> DELETED_LINK
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4867965570)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |
