@who:user
@what:links
@why:cleanliness

Feature: 03.1.ReplaceMyselfLinks

  Background:
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    And режим "replace-myself-links: run=true"
    And исходное сообщение со ссылкой на своё сообщение

  Scenario Outline: Ссылки на свои сообщения заменяются
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>)
    When пользователь отправляет исходное сообщение
    Then сообщение появляется в целевом чате
    And сообщение с замененной ссылкой

    Examples:
      | src_chat_id    | src_chat_name |
      | -1002641439846 | SRC PUB CH 1  |
      | -1002748936346 | SRC PUB CH 2  |
      | -1002792282007 | SRC PRV CH 1  |
      | -1002524362679 | SRC PRV CH 2  |
      | -1002736661856 | SRC PUB GRP 1 |
      | -1002781642357 | SRC PUB GRP 2 |
