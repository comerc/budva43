@who:user
@what:album_copy-once
@why:uniqueness

Feature: 12.MediaAlbumCopyOnce

  Background:
    Given исходный чат "-1002641439846" (SRC PUB CH 1)

  Scenario Outline: Медиаальбом копируется только один раз
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>) с copy-once
    When пользователь отправляет медиаальбом в исходный чат
    Then альбом появляется в целевом чате только один раз

    Examples:
      | dst_chat_id    | dst_chat_name |
      | -1002667730628 | DST PUB CH 1  |
      | -1002877966922 | DST PUB CH 2  |