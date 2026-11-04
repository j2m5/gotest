-- First neutral-world batch. Ownership remains in seed07; content grants no access.
-- Early Great Galactic War: authored districts/roles, informed by planetary lore.
-- No later Belsavis breakout, Taris reconstruction or Hoth starship graveyard.
-- Manaan routes use platforms and sealed rooms, not an underwater movement system.
-- Onderon does not include Dxun; Tatooine uses early settlements, not film-era towns.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('belsavis_fugitive','Вооружённый беглец','Скрывается в служебных постройках и тёплых разломах Белсависа. Отнимает припасы у проходящих экспедиций и защищает своё укрытие.','humanoid'),
 ('belsavis_rift_stalker','Хищник термального разлома','Обитает среди густых зарослей, согретых подземным теплом. Нападает из укрытий у ручьёв и узких проходов между ледяными стенами.','beast'),
 ('belsavis_service_droid','Сбойный тюремный техдроид','Обслуживал внешние хозяйственные узлы и линии теплоснабжения. Повреждённая программа заставляет его атаковать движущиеся объекты ремонтными инструментами.','droid'),
 ('belsavis_ancient_warden','Древний страж хранилища','Продолжает охранять подходы к забытым помещениям под поверхностью Белсависа. Распознаёт любого посетителя как нарушителя давно утраченных правил допуска.','droid'),
 ('bothawui_data_thief','Похититель ботавских данных','Крадёт закрытые сведения и оборудование у торговых домов Ботавуи. Защищает подпольные передатчики и тайники с похищенными носителями.','humanoid'),
 ('bothawui_forest_stalker','Хищник ботавских лесов','Держится в густых лесах и на каменистых окраинах равнин. Подстерегает добычу возле звериных троп и заброшенных стоянок.','beast'),
 ('bothawui_cargo_raider','Налётчик торговых окраин','Грабит караваны вдали от охраняемых терминалов Ботавуи. Использует пустующие склады для хранения добычи и подготовки новых нападений.','humanoid'),
 ('bothawui_sentry_droid','Перепрошитый охранник архива','Захваченная преступниками машина охраняет краденые сведения. Пропускает только участников своей группы и открывает огонь по остальным.','droid'),
 ('dantooine_kath_hound','Дантуинская кат-гончая','Рогатый зверь бродит по травянистым равнинам и холмам Дантуина. Защищает свою территорию у троп, водопоев и покинутых ферм.','beast'),
 ('dantooine_kinrath','Кинрат','Хищное членистоногое скрывается в дантуинских пещерах и зарослях. Нападает на существ, приблизившихся к гнезду или узким подземным проходам.','beast'),
 ('dantooine_raider','Налётчик дантуинских равнин','Грабит отдалённые хозяйства и небольшие грузовые караваны. Прячет припасы в заброшенных усадьбах и старых каменных постройках.','humanoid'),
 ('dantooine_harvest_droid','Неисправный сельскохозяйственный дроид','Остался без обслуживания на покинутой ферме. Принимает живых существ за препятствия и применяет тяжёлые рабочие инструменты.','droid'),
 ('taris_rakghoul','Тарисский ракгул','Обитает в заросших развалинах и тёмных переходах Тариса. Выслеживает добычу среди рухнувших перекрытий и заболоченных нижних кварталов.','beast'),
 ('taris_scrap_raider','Налётчик тарисских руин','Отбирает находки у сборщиков металла и исследователей. Охраняет склады награбленного на сохранившихся площадках разрушенного города.','humanoid'),
 ('taris_salvage_droid','Повреждённый разборочный дроид','Разбирал завалы и вытаскивал уцелевшие городские узлы. После отказа датчиков использует режущие инструменты против всего, что движется рядом.','droid'),
 ('taris_security_droid','Старый охранный автомат','Сохранился в закрытом тарисском комплексе после разрушения города. Продолжает защищать пустые помещения, не принимая современные коды доступа.','droid'),
 ('hoth_ice_stalker','Хищник ледяных расселин','Прячется в снежных ложбинах и у защищённых от ветра скал Хота. Подбирается к добыче вдоль ледяных уступов и занесённых троп.','beast'),
 ('hoth_wampa','Вампа','Крупный хищник занимает ледяные пещеры и охотится на открытых снежных просторах. Яростно защищает логово от любого приблизившегося существа.','beast'),
 ('hoth_expedition_raider','Грабитель полярных экспедиций','Нападает на малые исследовательские группы и перевозчиков снабжения. Укрывается в заброшенных полевых станциях, присваивая топливо и оборудование.','humanoid'),
 ('hoth_survey_droid','Сбойный полярный техдроид','Обслуживал разведочные площадки и буровые установки на Хоте. Повреждённая система наведения превращает его рабочие инструменты в опасное оружие.','droid'),
 ('tatooine_womp_rat','Вомп-крыса','Кормится среди пустынных камней, старых труб и брошенного оборудования. Нападает на тех, кто подходит слишком близко к её укрытию.','beast'),
 ('tatooine_tusken_raider','Тускенский налётчик','Участник вооружённой группы, устраивающей засады на пустынных дорогах. Охраняет занятые своей группой проходы и стоянки.','humanoid'),
 ('tatooine_scrap_droid','Неисправный дроид-сборщик','Собран из деталей списанной горной и грузовой техники. Сбойные датчики заставляют его разбирать всё, что попадает в зону захватов.','droid'),
 ('tatooine_desert_outlaw','Пустынный разбойник','Грабит перевозчиков воды, руды и запасных частей на Татуине. Держит краденые грузы в покинутых шахтах и укрытиях между скалами.','humanoid'),
 ('manaan_cargo_raider','Налётчик морских платформ','Захватывает грузы на удалённых причалах Манаана. Прячет добычу в закрытых отсеках и угрожает оружием свидетелям налётов.','humanoid'),
 ('manaan_kolto_smuggler','Контрабандист колто','Вывозит похищенные партии колто через заброшенные погрузочные площадки. Вооружённо защищает свои склады и поддельные грузовые документы.','humanoid'),
 ('manaan_service_droid','Сбойный платформенный техдроид','Обслуживал насосы, подъёмники и герметичные переходы. Повреждённые инструкции заставляют его устранять посетителей как помехи работе оборудования.','droid'),
 ('manaan_security_droid','Взломанный охранник платформы','Перепрограммирован контрабандистами для защиты нелегальных складов. Патрулирует сухие галереи и площадки над океаном, атакует посторонних.','droid'),
 ('onderon_cannok','Ондеронский каннок','Небольшой всеядный зверь разыскивает пищу в джунглях и на окраинах поселений. Может напасть, защищая добычу или привычное укрытие.','beast'),
 ('onderon_boma','Бома','Крупный зверь прокладывает себе путь сквозь ондеронские заросли. Защищает территорию и бросается на чужаков у водопоев и скальных логовищ.','beast'),
 ('onderon_beast_poacher','Ондеронский браконьер','Ловит диких животных для незаконной продажи. Охраняет ловушки, грузовые загоны и скрытые лесные стоянки.','humanoid'),
 ('onderon_rogue_handler','Преступный торговец зверями','Покупает добычу браконьеров и готовит животных к тайной перевозке. Вооружённо защищает перевалочные базы и помещения с контрабандным снаряжением.','humanoid')
ON CONFLICT (code) DO NOTHING;

-- Belsavis and Bothawui: outer prison infrastructure, thermal biomes and trade routes.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('belsavis','outer_service_landing','Внешняя служебная площадка','Грузовой настил расположен у ледяной стены вдали от действующих тюремных корпусов. Пустующие хозяйственные строения привлекают беглецов, ищущих припасы.',1,1),
 ('belsavis','ice_rift_approach','Подступы к тёплому разлому','Тропа спускается между ледяными уступами к полосе открытой почвы. Пар и камни скрывают брошенные приборы и небольшие убежища.',2,6),
 ('belsavis','thermal_grove','Роща термальных источников','Подземное тепло питает густую растительность посреди ледяного мира. В зарослях у ручьёв охотятся местные звери и скрываются вооружённые беглецы.',3,11),
 ('belsavis','abandoned_supply_store','Покинутые склады снабжения','Старые кладовые отделены от действующей зоны обрушившейся транспортной линией. Часть припасов присвоили беглецы, а неисправные машины продолжают обходить дворы.',4,16),
 ('belsavis','prison_utility_tunnels','Технические тоннели внешнего сектора','Тёплые сухие проходы соединяют заброшенные хозяйственные узлы. Повреждённые техдроиды и вооружённые обитатели укрытий контролируют тесные развилки.',5,21),
 ('belsavis','old_geothermal_station','Старая геотермальная станция','Каменные террасы окружают закрытый узел сбора подземного тепла. Его площадки заняты беглецами, пытающимися использовать оставшееся оборудование.',6,26),
 ('belsavis','overgrown_sinkhole','Заросшая котловина','За стеной зелени открываются древние каменные основания. Звери занимают тёплые впадины, а старые охранные машины перекрывают проходы к сооружениям.',7,31),
 ('belsavis','buried_guard_post','Погребённый сторожевой узел','Полузанесённые коридоры ведут к защитным помещениям неизвестного возраста. Активные стражи охраняют их наравне с давно опустевшими камерами.',8,36),
 ('belsavis','ancient_outer_vaults','Внешние галереи древних хранилищ','Массивные стены отделяют доступные проходы от запечатанных глубин. Исследователи редко уходят дальше первых залов, где до сих пор действуют автоматы.',9,41),
 ('belsavis','sealed_warden_complex','Забытый комплекс стражей','В глубине древнего сооружения сохранились посты охраны и узкие обходные галереи. Здесь сосредоточены наиболее опасные машины внешнего пояса хранилищ.',10,46),
 ('bothawui','outer_trade_landing','Окраина торгового порта','За пределами охраняемых причалов тянутся старые погрузочные дворы. Налётчики прячут здесь добычу и следят за небольшими грузовыми караванами.',1,1),
 ('bothawui','relay_service_yards','Служебные дворы ретрансляторов','Кабельные каналы и подсобные помещения окружают городские узлы связи. Похитители данных используют закрытые участки для подключения нелегального оборудования.',2,6),
 ('bothawui','woodland_caravan_paths','Лесные караванные тропы','Дороги уходят от торговых кварталов к зелёным равнинам. В зарослях скрываются хищники и грабители, поджидающие отставшие грузы.',3,11),
 ('bothawui','stolen_cargo_depots','Склады похищенных грузов','Пустующий складской ряд превращён в перевалочный пункт преступной группы. Между контейнерами ходят вооружённые охранники и взломанные автоматы.',4,16),
 ('bothawui','plains_convoy_route','Конвойный путь равнин','Грузовая дорога проходит через открытые поля и редкие рощи. Налётчики заняли старые стоянки, откуда удобно следить за движением перевозчиков.',5,21),
 ('bothawui','illicit_data_exchange','Подпольный центр обмена данными','В закрытых кабинетах торгового комплекса перепродают украденные сведения. Доступ к архивам охраняют преступники и перепрошитые машины.',6,26),
 ('bothawui','forested_ridge_pass','Перевал лесистого хребта','Тропа между скалами связывает равнины с удалёнными наблюдательными постами. Хищники и скрывающиеся от патрулей налётчики занимают защищённые уступы.',7,31),
 ('bothawui','seized_signal_station','Захваченная сигнальная станция','Брошенный передающий узел снова работает в интересах похитителей данных. В служебных залах хранят оборудование перехвата и краденые носители.',8,36),
 ('bothawui','hidden_archive_vaults','Тайные архивные хранилища','Укреплённые комнаты скрыты под пустующим деловым корпусом. Преступная сеть держит здесь ценные сведения и охраняет каждый переход.',9,41),
 ('bothawui','shadow_trade_compound','Убежище теневой торговой сети','Изолированный комплекс объединяет склады, передатчики и закрытые архивы. Его обитатели вооружили охрану и перекрыли подходы к главным помещениям.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Dantooine and Taris: grasslands and caves; ruins predating the current war.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('dantooine','farmstead_landing','Фермерская посадочная площадка','Небольшой настил стоит у края полей и пустующих хозяйственных сараев. Отсюда начинаются дороги к дальним фермам, где появляются налётчики и дикие звери.',1,1),
 ('dantooine','blba_grove_path','Тропа рощи блба','Кривые стволы блба возвышаются над высокой травой. Среди корней проходят звериные тропы, а у старых оград остаётся брошенная техника.',2,6),
 ('dantooine','grassland_caravan_road','Караванная дорога равнин','Пыльная колея пересекает холмы и обширные луга. Кат-гончие держатся у водопоев, а налётчики наблюдают за дорогой из покинутых навесов.',3,11),
 ('dantooine','abandoned_farmyards','Дворы покинутых ферм','Хозяйственные постройки постепенно зарастают травой. Неисправные сельскохозяйственные машины продолжают работать рядом с тайниками разбойников.',4,16),
 ('dantooine','rocky_ravine','Каменистый овраг','Вода прорезала узкий путь между обнажёнными скалами. Боковые расщелины заняты кинратами, выходящими к тропе из неглубоких пещер.',5,21),
 ('dantooine','old_grain_station','Старая зерновая станция','Рядом с пустыми зернохранилищами сохранились погрузочные площадки. Налётчики используют их для сортировки добычи и ремонта захваченных машин.',6,26),
 ('dantooine','crystal_cave_approach','Подступы к кристальным пещерам','Каменный спуск ведёт под холм, где в стенах мерцают минеральные жилы. Кинраты занимают тёмные ответвления и охраняют подземные гнёзда.',7,31),
 ('dantooine','ancient_stone_courts','Древние каменные дворы','На заросшем склоне сохранились основания старых сооружений. Разбойники устроили укрытия среди стен, а звери используют обвалы как логовища.',8,36),
 ('dantooine','deep_crystal_galleries','Глубокие кристальные галереи','Подземные коридоры сужаются между крупными наростами минералов. Здесь находится большое гнездовье кинратов, защищённое каменными разломами.',9,41),
 ('dantooine','raider_hill_fort','Укрепление налётчиков на холме','Покинутая усадьба превращена в защищённую базу грабителей равнин. Захваченная техника и вооружённые дозоры перекрывают дорогу к главным складам.',10,46),
 ('taris','raised_ruin_landing','Площадка над руинами','Уцелевший участок верхнего перекрытия пригоден для посадки малого транспорта. В соседних развалинах сборщики металла сталкиваются с грабителями и сбойной техникой.',1,1),
 ('taris','overgrown_causeway','Заросшая городская эстакада','Растительность оплетает опоры разрушенной магистрали. На сохранившихся пролётах видны следы разборочных машин и стоянок налётчиков.',2,6),
 ('taris','marsh_district_paths','Тропы заболоченного квартала','Сухие полосы старого дорожного покрытия проходят между топями и развалинами. Из тёмных проёмов к тропам выходят ракгулы.',3,11),
 ('taris','transit_station_ruins','Развалины пересадочной станции','Под рухнувшим сводом остались платформы и служебные переходы. Налётчики обыскивают транспортные узлы, не всегда замечая обитателей нижних ярусов.',4,16),
 ('taris','collapsed_cargo_blocks','Обрушенные грузовые корпуса','Погрузочные помещения завалены остатками башен и перекрытий. Между контейнерами ходят повреждённые разборочные машины и охраняющие добычу грабители.',5,21),
 ('taris','abandoned_pumping_station','Заброшенная насосная станция','Машинные залы возвышаются над затопленными техническими уровнями. Сухие обходные галереи заняты старыми автоматами и крадущимися среди труб ракгулами.',6,26),
 ('taris','lower_city_overpass','Переходы нижнего города','Сохранившиеся мосты соединяют тёмные остатки жилых и складских кварталов. Здесь редко встречаются исследователи, зато часто слышно движение местных хищников.',7,31),
 ('taris','buried_service_halls','Погребённые служебные залы','Под толщей обломков сохранились длинные коридоры обслуживания. Старые охранные автоматы защищают закрытые секции рядом с занятыми ракгулами провалами.',8,36),
 ('taris','old_civic_vaults','Хранилища старого городского комплекса','Толстые стены уберегли часть подземных помещений от разрушения. Теперь их двери охраняют уцелевшие машины, а проходы к ним заняты опасными обитателями руин.',9,41),
 ('taris','rakghoul_hollows','Логовища под разрушенными башнями','Обвалы образовали целую сеть сухих полостей над заболоченными глубинами. Ракгулы освоили эти укрытия и защищают узкие подходы к своим логовищам.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Hoth and Tatooine: small expeditions, native fauna and early desert outposts.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('hoth','icefield_landing','Посадочная площадка ледяной равнины','Защитные ограждения обозначают твёрдый участок среди снежных наносов. У пустующих грузовых навесов появляются грабители экспедиций и неисправные техдроиды.',1,1),
 ('hoth','snowbound_pass','Заснеженный перевал','Узкий проход между скалами позволяет обойти глубокие ледяные трещины. Хищники используют снежные карманы для укрытия, а налётчики следят за путниками.',2,6),
 ('hoth','old_survey_camp','Старый геодезический лагерь','Несколько модульных построек стоят вокруг брошенных измерительных опор. Запасы топлива и сохранившиеся приборы привлекают вооружённых мародёров.',3,11),
 ('hoth','frozen_shelf_trail','Тропа ледяного уступа','Ветер очистил узкую полосу вдоль высокого ледяного края. Среди скал скрываются местные хищники и оставленные экспедициями машины.',4,16),
 ('hoth','thermal_crevasse','Термальная расселина','Из глубины разлома поднимается тепло, оставляя доступными каменные выступы. В защищённых нишах устраиваются звери и грабители, избегающие открытого льда.',5,21),
 ('hoth','raider_supply_camp','Лагерь похитителей снабжения','За снежным валом спрятаны обогреваемые укрытия и краденые грузовые контейнеры. Налётчики охраняют топливо и переделанные рабочие машины.',6,26),
 ('hoth','blue_ice_caverns','Пещеры голубого льда','Толстые ледяные стены приглушают свет и звуки с поверхности. В глубине проходов вампы устраивают логовища и защищают остатки добычи.',7,31),
 ('hoth','abandoned_drilling_station','Покинутая буровая станция','Промёрзшие переходы соединяют жилые модули и старые буровые площадки. Часть помещений занята грабителями, а техдроиды продолжают опасные рабочие циклы.',8,36),
 ('hoth','glacial_gorge','Глубокое ледниковое ущелье','Узкие каменные полки проходят между высокими стенами льда. Местные хищники занимают боковые пещеры и подстерегают добычу на поворотах.',9,41),
 ('hoth','wampa_hunting_grounds','Охотничьи угодья вамп','Цепь крупных пещер выходит к закрытой от ветра снежной долине. Здесь держатся самые опасные хищники окрестностей, охраняющие свои логовища.',10,46),
 ('tatooine','anchorhead_outskirts','Окраины Анкорхеда','Пыльные хозяйственные площадки переходят в каменистую пустыню. Среди списанных агрегатов прячутся вомп-крысы и разбойники, высматривающие грузы.',1,1),
 ('tatooine','salvage_flats','Пустошь брошенной техники','Песок заносит остатки горных машин и разобранных грузовых платформ. Неисправные дроиды продолжают собирать детали рядом с тайниками мародёров.',2,6),
 ('tatooine','dune_caravan_trail','Караванная тропа дюн','Путь между барханами связывает поселения с дальними добывающими площадками. Вооружённые налётчики устраивают засады за гребнями песка.',3,11),
 ('tatooine','old_mine_yards','Дворы старых рудников','Подъёмные опоры возвышаются над пустыми приёмниками руды. Разбойники занимают уцелевшие склады, а сборочные дроиды бродят между отвалами.',4,16),
 ('tatooine','abandoned_moisture_station','Покинутая водосборная станция','Ржавые конденсаторы окружают низкие служебные постройки. Налётчики присвоили остатки запасов, вытеснив зверей к краям ограждённого двора.',5,21),
 ('tatooine','rocky_canyon_approach','Подступы к скальному каньону','Каменные проходы сужаются между обрывистыми стенами. Тускенская группа заняла удобные позиции над дорогой и охраняет свои стоянки.',6,26),
 ('tatooine','sandstone_caverns','Пещеры песчаника','Под скальными навесами скрываются прохладные залы и узкие переходы. Здесь гнездятся пустынные звери и прячут добычу участники вооружённых банд.',7,31),
 ('tatooine','raider_canyon_camp','Стоянка налётчиков в каньоне','Несколько укрытий соединены тропами по каменным уступам. Занявшая каньон тускенская группа охраняет входы и захваченные припасы.',8,36),
 ('tatooine','deep_mine_workings','Глубокие горные выработки','Заброшенные штольни уходят под старые отвалы. Разбойники устроили здесь склад оружия и запустили собранные из обломков рабочие машины.',9,41),
 ('tatooine','hidden_mesa_outpost','Тайная база на каменном плато','Укреплённые строения скрыты за высокими краями плато. Отсюда крупная банда направляет налёты на пустынные дороги и охраняет основные запасы добычи.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Manaan stays walkable; Onderon's jungle threats are not the whole Beast Rider culture.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('manaan','outer_platform_landing','Внешняя посадочная платформа','Грузовой настил поднят над океаном у закрытого торгового узла. В пустующих складах скрываются похитители грузов, а вдоль ограждений движутся сбойные машины.',1,1),
 ('manaan','loading_pier_walkways','Настилы погрузочных причалов','Сеть широких мостков соединяет склады и места стоянки грузовых судов. Налётчики заняли редко используемые участки, где удобно перегружать добычу.',2,6),
 ('manaan','closed_kolto_depots','Закрытые склады колто','Уцелевшие контейнерные залы примыкают к старому экспортному причалу. Контрабандисты хранят здесь украденное сырьё и поддельные накладные.',3,11),
 ('manaan','platform_service_bridges','Служебные мосты платформ','Прочные настилы проходят рядом с трубопроводами и подъёмниками. Неисправные техдроиды обслуживают оборудование и не пропускают живых посетителей.',4,16),
 ('manaan','idle_processing_halls','Пустующие перерабатывающие залы','Часть оборудования давно остановлена, но сухие производственные помещения сохранились. Контрабандисты используют их для переупаковки и сокрытия похищенных партий колто.',5,21),
 ('manaan','seized_cargo_berths','Захваченные грузовые стоянки','Закрытые причалы стали перевалочным пунктом вооружённой группы. Между швартовыми опорами и контейнерами ходят налётчики и взломанные охранные машины.',6,26),
 ('manaan','sealed_lift_station','Герметичная подъёмная станция','Лифты соединяют поверхность с сухими помещениями внутри опор платформы. Служебные коридоры заняты техдроидами и охраной контрабандных складов.',7,31),
 ('manaan','lower_dry_galleries','Нижние герметичные галереи','За толстыми стенами шумит океан, но внутренние проходы остаются сухими. В закрытых отсеках устроены тайники контрабандистов и посты охранных автоматов.',8,36),
 ('manaan','old_pump_control','Старый центр управления насосами','Просторные герметичные залы окружают пульты заброшенного технического узла. Преступная группа восстановила часть оборудования и укрепила подходы к складам.',9,41),
 ('manaan','hidden_export_hub','Тайный узел вывоза колто','Закрытый комплекс связан с поверхностью защищёнными подъёмниками и сухими переходами. Здесь сосредоточены похищенные партии сырья и главные силы охраны контрабандистов.',10,46),
 ('onderon','iziz_outer_landing','Внешняя площадка Изиза','Небольшая посадочная зона расположена у дороги за городскими укреплениями. В соседних хозяйственных постройках появляются похитители зверей и дикие канноки.',1,1),
 ('onderon','field_edge_paths','Тропы у края полей','Открытые участки сменяются густой растительностью. Канноки выходят искать корм, а браконьеры расставляют ловушки вдали от людных дорог.',2,6),
 ('onderon','forest_stream_crossing','Переправа лесного ручья','Каменные отмели позволяют перейти поток между заросшими берегами. Звери держатся у воды, привлекая охотящихся за ними преступников.',3,11),
 ('onderon','jungle_canopy_trail','Тропа под пологом джунглей','Высокие деревья закрывают небо, оставляя узкую полосу прохода среди корней. На развилках встречаются звериные укрытия и скрытые стоянки браконьеров.',4,16),
 ('onderon','poacher_supply_yard','Склад браконьерского снабжения','В старом лесном дворе собраны ловушки, клетки и перевозочное снаряжение. Вооружённые торговцы охраняют припасы и готовят очередной вывоз животных.',5,21),
 ('onderon','rocky_jungle_ridge','Скалистый гребень джунглей','Тропа поднимается над густой зеленью к каменным уступам. Браконьеры заняли несколько защищённых площадок, откуда наблюдают за звериными переходами.',6,26),
 ('onderon','ruined_hunting_post','Развалины лесного поста','Старые каменные стены почти скрыты растительностью. Вокруг руин бродят бомы, а торговцы зверями используют сохранившиеся помещения как укрытия.',7,31),
 ('onderon','basalt_gorge','Базальтовое ущелье','Между тёмными скалами проходят узкие тропы к глубоким участкам леса. Крупные звери занимают боковые впадины, где браконьеры пытаются устроить засады.',8,36),
 ('onderon','illegal_beast_enclosures','Тайные загоны торговцев зверями','Укреплённые ограждения скрыты в удалённой лесной котловине. Преступники собрали здесь пойманных животных и защищают площадки подготовки к перевозке.',9,41),
 ('onderon','deep_jungle_dens','Логовища глубоких джунглей','Древние деревья и каменные арки окружают цепь крупных природных укрытий. Бомы и другие звери защищают свои территории далеко от городских дорог.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('belsavis','outer_service_landing','belsavis_fugitive','belsavis_service_droid','belsavis_fugitive'),
 ('belsavis','ice_rift_approach','belsavis_fugitive','belsavis_service_droid','belsavis_rift_stalker'),
 ('belsavis','thermal_grove','belsavis_rift_stalker','belsavis_fugitive','belsavis_rift_stalker'),
 ('belsavis','abandoned_supply_store','belsavis_fugitive','belsavis_service_droid','belsavis_fugitive'),
 ('belsavis','prison_utility_tunnels','belsavis_service_droid','belsavis_fugitive','belsavis_service_droid'),
 ('belsavis','old_geothermal_station','belsavis_fugitive','belsavis_service_droid','belsavis_fugitive'),
 ('belsavis','overgrown_sinkhole','belsavis_rift_stalker','belsavis_ancient_warden','belsavis_rift_stalker'),
 ('belsavis','buried_guard_post','belsavis_ancient_warden','belsavis_ancient_warden','belsavis_ancient_warden'),
 ('belsavis','ancient_outer_vaults','belsavis_ancient_warden','belsavis_ancient_warden','belsavis_ancient_warden'),
 ('belsavis','sealed_warden_complex','belsavis_ancient_warden','belsavis_ancient_warden','belsavis_ancient_warden'),
 ('bothawui','outer_trade_landing','bothawui_cargo_raider','rogue_loader','bothawui_cargo_raider'),
 ('bothawui','relay_service_yards','bothawui_data_thief','bothawui_sentry_droid','bothawui_data_thief'),
 ('bothawui','woodland_caravan_paths','bothawui_forest_stalker','bothawui_cargo_raider','bothawui_forest_stalker'),
 ('bothawui','stolen_cargo_depots','bothawui_cargo_raider','bothawui_sentry_droid','bothawui_cargo_raider'),
 ('bothawui','plains_convoy_route','bothawui_cargo_raider','bothawui_forest_stalker','bothawui_cargo_raider'),
 ('bothawui','illicit_data_exchange','bothawui_data_thief','bothawui_sentry_droid','bothawui_data_thief'),
 ('bothawui','forested_ridge_pass','bothawui_forest_stalker','bothawui_cargo_raider','bothawui_forest_stalker'),
 ('bothawui','seized_signal_station','bothawui_data_thief','bothawui_sentry_droid','bothawui_sentry_droid'),
 ('bothawui','hidden_archive_vaults','bothawui_data_thief','bothawui_sentry_droid','bothawui_sentry_droid'),
 ('bothawui','shadow_trade_compound','bothawui_cargo_raider','bothawui_data_thief','bothawui_sentry_droid'),
 ('dantooine','farmstead_landing','dantooine_raider','dantooine_harvest_droid','dantooine_kath_hound'),
 ('dantooine','blba_grove_path','dantooine_kath_hound','dantooine_harvest_droid','dantooine_kath_hound'),
 ('dantooine','grassland_caravan_road','dantooine_kath_hound','dantooine_raider','dantooine_raider'),
 ('dantooine','abandoned_farmyards','dantooine_harvest_droid','dantooine_raider','dantooine_harvest_droid'),
 ('dantooine','rocky_ravine','dantooine_kinrath','dantooine_kinrath','dantooine_kinrath'),
 ('dantooine','old_grain_station','dantooine_raider','dantooine_harvest_droid','dantooine_raider'),
 ('dantooine','crystal_cave_approach','dantooine_kinrath','dantooine_kinrath','dantooine_kinrath'),
 ('dantooine','ancient_stone_courts','dantooine_raider','dantooine_kath_hound','dantooine_raider'),
 ('dantooine','deep_crystal_galleries','dantooine_kinrath','dantooine_kinrath','dantooine_kinrath'),
 ('dantooine','raider_hill_fort','dantooine_raider','dantooine_harvest_droid','dantooine_raider'),
 ('taris','raised_ruin_landing','taris_scrap_raider','taris_salvage_droid','taris_scrap_raider'),
 ('taris','overgrown_causeway','taris_salvage_droid','taris_scrap_raider','taris_salvage_droid'),
 ('taris','marsh_district_paths','taris_rakghoul','taris_rakghoul','taris_rakghoul'),
 ('taris','transit_station_ruins','taris_scrap_raider','taris_rakghoul','taris_scrap_raider'),
 ('taris','collapsed_cargo_blocks','taris_salvage_droid','taris_scrap_raider','taris_salvage_droid'),
 ('taris','abandoned_pumping_station','taris_salvage_droid','taris_rakghoul','taris_security_droid'),
 ('taris','lower_city_overpass','taris_rakghoul','taris_rakghoul','taris_rakghoul'),
 ('taris','buried_service_halls','taris_rakghoul','taris_security_droid','taris_security_droid'),
 ('taris','old_civic_vaults','taris_rakghoul','taris_security_droid','taris_security_droid'),
 ('taris','rakghoul_hollows','taris_rakghoul','taris_rakghoul','taris_rakghoul'),
 ('hoth','icefield_landing','hoth_expedition_raider','hoth_survey_droid','hoth_expedition_raider'),
 ('hoth','snowbound_pass','hoth_ice_stalker','hoth_expedition_raider','hoth_ice_stalker'),
 ('hoth','old_survey_camp','hoth_expedition_raider','hoth_survey_droid','hoth_expedition_raider'),
 ('hoth','frozen_shelf_trail','hoth_ice_stalker','hoth_survey_droid','hoth_ice_stalker'),
 ('hoth','thermal_crevasse','hoth_ice_stalker','hoth_expedition_raider','hoth_ice_stalker'),
 ('hoth','raider_supply_camp','hoth_expedition_raider','hoth_survey_droid','hoth_expedition_raider'),
 ('hoth','blue_ice_caverns','hoth_ice_stalker','hoth_wampa','hoth_wampa'),
 ('hoth','abandoned_drilling_station','hoth_expedition_raider','hoth_survey_droid','hoth_survey_droid'),
 ('hoth','glacial_gorge','hoth_ice_stalker','hoth_wampa','hoth_wampa'),
 ('hoth','wampa_hunting_grounds','hoth_ice_stalker','hoth_wampa','hoth_wampa'),
 ('tatooine','anchorhead_outskirts','tatooine_womp_rat','tatooine_desert_outlaw','tatooine_desert_outlaw'),
 ('tatooine','salvage_flats','tatooine_scrap_droid','tatooine_desert_outlaw','tatooine_scrap_droid'),
 ('tatooine','dune_caravan_trail','tatooine_desert_outlaw','tatooine_tusken_raider','tatooine_tusken_raider'),
 ('tatooine','old_mine_yards','tatooine_desert_outlaw','tatooine_scrap_droid','tatooine_desert_outlaw'),
 ('tatooine','abandoned_moisture_station','tatooine_womp_rat','tatooine_desert_outlaw','tatooine_desert_outlaw'),
 ('tatooine','rocky_canyon_approach','tatooine_tusken_raider','tatooine_tusken_raider','tatooine_tusken_raider'),
 ('tatooine','sandstone_caverns','tatooine_womp_rat','tatooine_desert_outlaw','tatooine_tusken_raider'),
 ('tatooine','raider_canyon_camp','tatooine_tusken_raider','tatooine_tusken_raider','tatooine_tusken_raider'),
 ('tatooine','deep_mine_workings','tatooine_scrap_droid','tatooine_desert_outlaw','tatooine_scrap_droid'),
 ('tatooine','hidden_mesa_outpost','tatooine_desert_outlaw','tatooine_scrap_droid','tatooine_desert_outlaw'),
 ('manaan','outer_platform_landing','manaan_cargo_raider','manaan_service_droid','manaan_cargo_raider'),
 ('manaan','loading_pier_walkways','manaan_cargo_raider','manaan_service_droid','manaan_cargo_raider'),
 ('manaan','closed_kolto_depots','manaan_kolto_smuggler','manaan_security_droid','manaan_kolto_smuggler'),
 ('manaan','platform_service_bridges','manaan_service_droid','manaan_service_droid','manaan_service_droid'),
 ('manaan','idle_processing_halls','manaan_kolto_smuggler','manaan_service_droid','manaan_kolto_smuggler'),
 ('manaan','seized_cargo_berths','manaan_cargo_raider','manaan_security_droid','manaan_security_droid'),
 ('manaan','sealed_lift_station','manaan_service_droid','manaan_kolto_smuggler','manaan_security_droid'),
 ('manaan','lower_dry_galleries','manaan_kolto_smuggler','manaan_security_droid','manaan_security_droid'),
 ('manaan','old_pump_control','manaan_service_droid','manaan_kolto_smuggler','manaan_security_droid'),
 ('manaan','hidden_export_hub','manaan_kolto_smuggler','manaan_security_droid','manaan_security_droid'),
 ('onderon','iziz_outer_landing','onderon_cannok','onderon_beast_poacher','onderon_beast_poacher'),
 ('onderon','field_edge_paths','onderon_cannok','onderon_beast_poacher','onderon_cannok'),
 ('onderon','forest_stream_crossing','onderon_cannok','onderon_beast_poacher','onderon_cannok'),
 ('onderon','jungle_canopy_trail','onderon_cannok','onderon_beast_poacher','onderon_beast_poacher'),
 ('onderon','poacher_supply_yard','onderon_beast_poacher','onderon_rogue_handler','onderon_rogue_handler'),
 ('onderon','rocky_jungle_ridge','onderon_cannok','onderon_beast_poacher','onderon_beast_poacher'),
 ('onderon','ruined_hunting_post','onderon_rogue_handler','onderon_boma','onderon_rogue_handler'),
 ('onderon','basalt_gorge','onderon_beast_poacher','onderon_boma','onderon_boma'),
 ('onderon','illegal_beast_enclosures','onderon_beast_poacher','onderon_rogue_handler','onderon_boma'),
 ('onderon','deep_jungle_dens','onderon_cannok','onderon_boma','onderon_boma')
) AS v(planet_code,location_code,normal_template,strong_template,elite_template)
JOIN planets p ON p.code=v.planet_code
JOIN locations l ON l.planet_id=p.id AND l.code=v.location_code
CROSS JOIN LATERAL (VALUES
 ('normal',1,0,'normal',v.normal_template),
 ('strong',2,2,'strong',v.strong_template),
 ('elite',3,4,'elite',v.elite_template)
) AS r(code,position,level_offset,rank,template_code)
JOIN npc_templates t ON t.code=r.template_code
ON CONFLICT (location_id,code) DO NOTHING;

INSERT INTO npc_spawns(location_npc_id,slot)
SELECT n.id,slots.slot FROM location_npcs n
JOIN locations l ON l.id=n.location_id
JOIN planets p ON p.id=l.planet_id
CROSS JOIN LATERAL generate_series(1,CASE n.code WHEN 'normal' THEN 2 ELSE 1 END) AS slots(slot)
WHERE p.code IN ('belsavis','bothawui','dantooine','taris','hoth','tatooine','manaan','onderon') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
