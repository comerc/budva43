@who:user
@what:sources-link-title
@why:identification

Feature: 09.SourcesLinkTitle

  Background:
    Given будет пересылка - копия
    And будет ссылка

  Scenario Outline: Вставка ссылки с заголовком источника
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
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
