@who:user
@what:links
@why:cleanliness

Feature: 03.2.delete_external_links

  Background:
    Given будет пересылка - копия

  Scenario Outline: Ссылки на внешние сообщения удаляются
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    When пользователь отправляет YETI_MESSAGE
    Then пауза 10 сек.
    And YETI_MESSAGE в чате
    # нет пересылки для YETI_MESSAGE
    Given сообщение со ссылкой на последнее сообщение
    And будет замена: ссылка на YETI_MESSAGE -> DELETED_LINK
    When пользователь отправляет сообщение
    Then пауза 10 сек.
    And сообщение в чате "-1002667730628" (DST PUB CHL 1)
    And сообщение в чате "-1002473038431" (DST PRV CHL 1)
    And сообщение в чате "-1002866470933" (DST PUB GRP 1)
    And сообщение в чате "-4867965570" (DST PRV GRP 1)

    Examples:
      | src_chat_id    | src_chat_name |
      | -1002641439846 | SRC PUB CHL 1 |
      | -1002792282007 | SRC PRV CHL 1 |
      | -1002736661856 | SRC PUB GRP 1 |
