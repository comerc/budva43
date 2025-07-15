@who:user
@what:links
@why:cleanliness

Feature: 03.2.DeleteExternalLinks

  Background:
    Given будет замена: ссылка на YETI_MESSAGE -> DELETED_LINK

  Scenario Outline: Ссылки на внешние сообщения удаляются
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    When пользователь отправляет YETI_MESSAGE в исходный чат
    Then пауза 10 сек.
    And YETI_MESSAGE в исходном чате
    Given сообщение со ссылкой на последнее сообщение
    When пользователь отправляет сообщение в исходный чат
    Then пауза 10 сек.
    And сообщение в чате "-1002667730628" (DST PUB CH 1)
    And сообщение в чате "-1002877966922" (DST PUB CH 2)
    And сообщение в чате "-1002473038431" (DST PRV CH 1)
    And сообщение в чате "-1002641980237" (DST PRV CH 2)
    And сообщение в чате "-1002866470933" (DST PUB GRP 1)
    And сообщение в чате "-1002876400294" (DST PUB GRP 2)
    And сообщение в чате "-4867965570" (DST GRP 1)
    And сообщение в чате "-4913098869" (DST GRP 2)

    Examples:
      | src_chat_id    | src_chat_name |
      | -1002641439846 | SRC PUB CH 1  |
      | -1002792282007 | SRC PRV CH 1  |
      | -1002736661856 | SRC PUB GRP 1 |
