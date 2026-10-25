INSERT INTO factions (code, name, alias, description, icon) VALUES
  (
      'sith_empire',
      'Sith Empire',
      'Империя',
      'Жёсткое милитаристское общество, где инопланетные расы покорены и Лорды Ситов являются высшими, Империей правит таинственный, всемогущий Император Ситов. Для граждан Империи, власть всё; только сильнейший может взойти к славе.',
      'images/factions/sith_empire.gif'
  ),
  (
      'galactic_republic',
      'Galactic Republic',
      'Республика',
      'Основана 20000 лет назад исследователями и дипломатами, Галактическая Республика является хаотичным объединением многих миров, корпораций и видов, которые часто сталкиваются друг с другом.',
      'images/factions/galactic_republic.gif'
  )
ON CONFLICT (code) DO NOTHING;
