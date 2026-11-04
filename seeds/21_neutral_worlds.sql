-- Fifth neutral-world batch. Ownership/access remain independent of content.
-- Early Great Galactic War: authored districts and roles, informed by Legends.
-- All Dorin/Gand areas are sealed ground facilities with visitor-compatible air;
-- the small Gand enclave and permitted Togorian field routes are game adaptations.
-- Taanab precedes its agricultural colonization; no later Bimm colonial cities.
-- Eres III's partial regrowth is an explicit game assumption, not a dated lore fact.
-- No peaceful native species as enemy categories, orbital areas or new hazard mechanics.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('dorin_cargo_raider','Налётчик закрытых грузовых секций','Приезжий преступник занял склады герметичного комплекса на Дорине. Грабит прибывающие грузы и держится в секторах с пригодной для него атмосферой.','humanoid'),
 ('dorin_service_droid','Сбойный дроид внутреннего обслуживания','Продолжает обходить закрытые служебные помещения. Повреждённое распознавание заставляет его применять рабочие инструменты против посетителей.','droid'),
 ('dorin_survey_droid','Повреждённый исследовательский дроид','Обслуживал приборы внутри изолированного наблюдательного комплекса. Ошибочные команды превратили машину в угрозу для персонала и приезжих.','droid'),
 ('dorin_outlaw_technician','Техник банды приезжих','Разбирает похищенное оборудование в герметичных мастерских. Переоборудует рабочие машины для охраны тайников своей группы.','humanoid'),
 ('glee_anselm_island_stalker','Хищник островных зарослей','Скрывается в густой растительности небольших островов Гли-Ансельма. Выслеживает добычу возле сухих троп и береговых уступов.','beast'),
 ('glee_anselm_tidal_crawler','Прибрежный панцирник','Живёт у приливных мелководий и способен выбираться на сушу. Защищает укрытия между камнями и опорами старых настилов.','beast'),
 ('glee_anselm_cove_raider','Налётчик островных бухт','Грабит перевозчиков и занимает удалённые причалы. Хранит добычу в покинутых береговых складах.','humanoid'),
 ('glee_anselm_pier_droid','Сбойный причальный дроид','Остался на старой погрузочной площадке. Изношенные датчики заставляют его хватать посетителей вместо грузов.','droid'),
 ('gand_expedition_raider','Грабитель экспедиционных секций','Преступник из числа приезжих захватил часть закрытого комплекса на Ганде. Держится в помещениях с привычной ему атмосферой и похищает оборудование.','humanoid'),
 ('gand_sample_droid','Сбойный лабораторный дроид','Обслуживал изолированные приборные камеры. Повреждённое распознавание заставляет его применять манипуляторы против посетителей.','droid'),
 ('gand_freight_droid','Повреждённый шлюзовой погрузчик','Перемещал контейнеры внутри герметичных грузовых секций. Из-за сбоя нападает на движущиеся объекты рабочими захватами.','droid'),
 ('gand_smuggler_guard','Охранник тайников приезжих','Вооружённый участник преступной группы охраняет украденные приборы. Удерживает внутренние коридоры с атмосферой для приезжих.','humanoid'),
 ('togoria_plains_stalker','Хищник травянистых ложбин','Прячется в высокой траве и у подножий холмов Тогории. Выслеживает добычу возле звериных троп и тенистых укрытий.','beast'),
 ('togoria_liphon','Липхон','Крылатая хищная рептилия Тогории. Нападает на добычу у открытых склонов и охраняет места отдыха среди скальных выступов.','beast'),
 ('togoria_poacher','Приезжий браконьер','Незаконно охотится на местных животных вдали от поселений. Скрывает снаряжение и захваченную добычу на временных стоянках.','humanoid'),
 ('togoria_survey_droid','Повреждённый дроид приезжих','Был оставлен на небольшом полевом посту. Сбойная программа заставляет машину атаковать тех, кто подходит к оборудованию.','droid'),
 ('eres_iii_plain_stalker','Хищник пустошей Эреса','Охотится среди каменных ложбин и редкой растительности Эреса III. Использует старые углубления в земле как укрытия.','beast'),
 ('eres_iii_burrower','Роющий зверь равнин','Устраивает норы в сухой почве между участками травы. Защищает входы в подземные ходы и ищет корм возле каменных гряд.','beast'),
 ('eres_iii_salvage_raider','Грабитель исследовательских стоянок','Похищает припасы и приборы экспедиций, изучающих следы старой войны. Держит тайники в покинутых полевых навесах.','humanoid'),
 ('eres_iii_survey_droid','Сбойный полевой исследователь','Привезён для обследования грунта и старых разрушений. Повреждённые датчики заставляют его применять инструменты против посетителей.','droid'),
 ('taanab_jungle_stalker','Хищник джунглей Таанаба','Выслеживает добычу под густым пологом неосвоенных лесов. Скрывается среди лиан, корней и каменных ниш.','beast'),
 ('taanab_root_crawler','Корневой ползун Таанаба','Обитает в подлеске и полостях между крупными корнями. Защищает свои укрытия от существ, проходящих по узким тропам.','beast'),
 ('taanab_expedition_raider','Грабитель джунглевых экспедиций','Перехватывает припасы небольших исследовательских групп. Устраивает тайники в заброшенных временных лагерях.','humanoid'),
 ('taanab_survey_droid','Сбойный разведчик джунглей','Остался на покинутом полевом посту. Сломанные датчики заставляют его считать движущиеся объекты помехами для обследования.','droid'),
 ('zeltros_cargo_thief','Похититель грузов Зелтроса','Вскрывает контейнеры в закрытых служебных кварталах. Прячет украденные товары в пустующих складах за пределами прогулочных улиц.','humanoid'),
 ('zeltros_gang_enforcer','Боевик тайных павильонов','Вооружённый преступник охраняет склады и мастерские своей группы. Не допускает посторонних в захваченные служебные помещения.','humanoid'),
 ('zeltros_stage_droid','Сбойный дроид декораций','Остался в закрытой мастерской оформления павильонов. Повреждённые команды заставляют его применять монтажные захваты против посетителей.','droid'),
 ('zeltros_service_droid','Повреждённый хозяйственный дроид','Продолжает обходить покинутые внутренние дворы. Ошибочно распознаёт движущихся существ как препятствия для уборки и погрузки.','droid'),
 ('cerea_grassland_stalker','Хищник травяных равнин','Обитает в кустарниковых ложбинах Цереи. Выслеживает добычу возле тихих троп вдали от обрабатываемых участков.','beast'),
 ('cerea_riverbank_predator','Хищник речных зарослей','Скрывается среди растительности на берегах рек. Выходит на сухие тропы и защищает логовища у воды.','beast'),
 ('cerea_outlander_raider','Грабитель из числа приезжих','Нападает на небольшие пункты снабжения и перевозчиков. Прячет добычу на заброшенных окраинных стоянках.','humanoid'),
 ('cerea_field_droid','Повреждённый привозной дроид','Был оставлен у полевого поста приезжих. Из-за сбоя защищает пустые грузовые площадки и атакует посетителей.','droid'),
 ('bimmisaari_forest_stalker','Хищник лесов асаари','Скрывается среди корней и густого подлеска Биммисаари. Выслеживает добычу у тенистых переходов между лесными полянами.','beast'),
 ('bimmisaari_thicket_crawler','Ползун влажного подлеска','Обитает в полостях под поваленными стволами и корнями. Защищает укрытия возле узких лесных троп.','beast'),
 ('bimmisaari_offworld_poacher','Приезжий лесной браконьер','Незаконно ловит животных вдали от местных поселений. Держит тайники в покинутых экспедиционных стоянках.','humanoid'),
 ('bimmisaari_survey_droid','Сбойный дроид лесной экспедиции','Остался возле временных приборных постов приезжих. Повреждённые инструкции заставляют его нападать на движущиеся объекты.','droid')
ON CONFLICT (code) DO NOTHING;

-- Dorin.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('dorin','sealed_arrival_hall','Герметичный зал прибытия','Закрытый посадочный ангар соединён шлюзами с комплексом для приезжих. Внутри поддерживается пригодный для них воздух, а пустующие грузовые ячейки заняли налётчики.',1,1),
 ('dorin','inner_transfer_passage','Внутренний перевалочный коридор','Герметичный переход соединяет ангар с хозяйственными секциями. Преступники перехватывают контейнеры у старых погрузочных постов.',2,6),
 ('dorin','closed_instrument_store','Закрытый склад приборов','Изолированные кладовые сохранили собственное жизнеобеспечение. Техники банды вскрывают ящики рядом с повреждёнными исследовательскими машинами.',3,11),
 ('dorin','protected_observation_gallery','Защищённая наблюдательная галерея','За прочным остеклением видны облака Дорина, но внутренний воздух отделён от внешней среды. Сбойные дроиды продолжают движение вдоль приборных стоек.',4,16),
 ('dorin','sealed_maintenance_annex','Герметичный служебный корпус','Внутренние мастерские соединены крытыми переходами со шлюзами. Налётчики заняли ремонтные помещения, сохранив исправную автоматику воздуха.',5,21),
 ('dorin','isolated_sensor_laboratory','Изолированная лаборатория датчиков','Лабораторные залы находятся внутри общего защитного контура комплекса. Повреждённые исследовательские машины обходят пустые рабочие места.',6,26),
 ('dorin','covered_cargo_locks','Крытые грузовые шлюзы','Цепочка закрытых камер позволяет перемещать контейнеры между внутренними секторами. Банда охраняет захваченные площадки с помощью переделанной техники.',7,31),
 ('dorin','outlaw_pressure_workshops','Герметичные мастерские банды','Укреплённые помещения скрыты за внутренними техническими дверями. Преступники ремонтируют здесь машины и сортируют похищенные приборы.',8,36),
 ('dorin','deep_service_sections','Глубокие служебные секции','Защитные стены отделяют внутренние склады от внешней атмосферы. Сбойные рабочие дроиды продолжают обходить узкие проходы.',9,41),
 ('dorin','fortified_guest_depot','Укреплённое хранилище комплекса','Внутри герметичного корпуса банда устроила основной склад. Налётчики и переделанные машины защищают его, не нарушая работающего жизнеобеспечения.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Glee Anselm.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('glee_anselm','sheltered_island_landing','Укрытая островная площадка','Небольшой посадочный настил примыкает к береговым складам. Налётчики заняли пустующие навесы рядом с оставленной техникой.',1,1),
 ('glee_anselm','coastal_thicket_path','Тропа прибрежных зарослей','Сухая дорожка проходит между густой зеленью и каменным берегом. Хищники скрываются в тени, а панцирники выходят из мелких заводей.',2,6),
 ('glee_anselm','abandoned_fishing_piers','Покинутые рыбацкие причалы','Прочные мостки соединяют пустые хозяйственные площадки над водой. Преступники используют навесы, пока сбойные дроиды обходят настилы.',3,11),
 ('glee_anselm','tidal_pool_walkway','Мостки приливных заводей','Приподнятый путь огибает мелкие водоёмы и мокрые камни. Панцирники занимают щели под опорами и выходы на сухой берег.',4,16),
 ('glee_anselm','wooded_island_ridge','Лесистый островной гребень','Невысокая гряда возвышается над заболоченной частью острова. В густых зарослях вдоль тропы прячутся крупные хищники.',5,21),
 ('glee_anselm','raider_boathouse_yards','Дворы захваченных лодочных сараев','Банда укрепила старые навесы у защищённой бухты. Похищенные грузы хранятся под охраной налётчиков и повреждённых машин.',6,26),
 ('glee_anselm','wetland_boardwalk','Настилы островных топей','Высокие мостки ведут над влажными низинами к сухим участкам берега. Панцирники цепляются за опоры, а звери наблюдают из прибрежных зарослей.',7,31),
 ('glee_anselm','stone_shore_den_path','Тропа береговых логовищ','Каменная полка соединяет сухие ниши вдоль берега. Прибрежные существа защищают укрытия под нависающими выступами.',8,36),
 ('glee_anselm','hidden_cove_storehouses','Скрытые склады бухты','Пустующие береговые постройки закрыты рядами украденных контейнеров. Налётчики охраняют тайники и старые механизмы погрузки.',9,41),
 ('glee_anselm','fortified_island_pier','Укреплённый островной причал','Банда превратила дальний сухопутный терминал в опорный пункт. Посты и переделанная техника перекрывают подходы со стороны береговой дороги.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Gand.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('gand','expedition_arrival_lock','Шлюз прибытия экспедиции','Герметичный ангар принадлежит небольшому наземному комплексу для допущенных приезжих. Внутри поддерживается привычный им воздух; часть грузовых секций заняли грабители.',1,1),
 ('gand','inner_sample_passage','Внутренний коридор образцов','Закрытый переход отделён от аммиачного тумана несколькими шлюзами. Сбойные машины продолжают двигаться между пустыми приёмными ячейками.',2,6),
 ('gand','sealed_supply_rooms','Герметичные кладовые снабжения','Небольшие склады находятся внутри жилого защитного контура. Преступники заняли дальние комнаты и охраняют похищенные приборы.',3,11),
 ('gand','protected_mist_observatory','Защищённый павильон наблюдения','Внешние туманы видны за плотным остеклением, внутренняя атмосфера изолирована. Повреждённые дроиды обходят оставленные приборные стойки.',4,16),
 ('gand','closed_loading_chambers','Закрытые погрузочные камеры','Последовательные шлюзы соединяют внутренние грузовые площадки комплекса. Сбойные погрузчики продолжают выполнять команды без операторов.',5,21),
 ('gand','seized_expedition_store','Захваченный экспедиционный склад','Грабители укрепили кладовые внутри герметичного служебного крыла. Они охраняют добычу, сохраняя рабочую систему воздуха.',6,26),
 ('gand','isolated_instrument_wing','Изолированное приборное крыло','Залы защищены от внешней атмосферы и соединены внутренними переходами. Лабораторные машины действуют по повреждённым инструкциям.',7,31),
 ('gand','sealed_return_galleries','Герметичные возвратные галереи','Крытые коридоры связывают дальние камеры со складами комплекса. Преступники используют их для скрытого перемещения украденного оборудования.',8,36),
 ('gand','outlaw_inner_workrooms','Внутренние мастерские преступников','Небольшие помещения переделаны в тайный пункт сортировки приборов. Вооружённая охрана и сбойные машины удерживают входы из герметичного коридора.',9,41),
 ('gand','fortified_expedition_vault','Укреплённое хранилище экспедиции','Защитные стены и исправные шлюзы отделяют склад от внешней среды. Банда держит здесь основные запасы, пользуясь автономным жизнеобеспечением.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Togoria.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('togoria','field_supply_landing','Полевая площадка снабжения','Небольшой пост стоит у начала маршрута для допущенных приезжих. Браконьеры заняли пустые навесы, где осталось привозное оборудование.',1,1),
 ('togoria','tall_grass_track','Тропа высокой травы','Узкая дорожка огибает травянистые холмы. Хищники подстерегают добычу в ложбинах, а липхоны наблюдают с каменных выступов.',2,6),
 ('togoria','woodland_edge_paths','Тропы лесной опушки','Открытая равнина переходит в густые тенистые заросли. Звери держатся возле боковых лазов и сухих корневых ниш.',3,11),
 ('togoria','abandoned_field_sheds','Покинутые полевые навесы','От временной стоянки приезжих остались лёгкие постройки и ящики. Браконьеры используют их рядом со сбойной разведочной машиной.',4,16),
 ('togoria','windy_hill_pass','Ветреный проход между холмами','Тропа проходит у скальных выступов над травяными низинами. Липхоны спускаются к земле в поисках добычи.',5,21),
 ('togoria','poacher_hidden_camp','Скрытый лагерь браконьеров','Густые заросли укрывают временные палатки и тайники снаряжения. Приезжая группа охраняет добычу и свои повреждённые механизмы.',6,26),
 ('togoria','deep_thicket_hollows','Глубокие ложбины зарослей','Холмы закрывают участок густой растительности от ветра. Местные хищники занимают сухие укрытия и выходы к открытой траве.',7,31),
 ('togoria','liphon_rock_ledges','Скальные уступы липхонов','Широкие каменные полки поднимаются над невысокой грядой. Крылатые хищники используют их для отдыха и охраняют подходы.',8,36),
 ('togoria','remote_poacher_store','Дальний тайник браконьеров','Ящики и переносные навесы скрыты между камнями. Группа приезжих удерживает запасы и оставленные у тропы машины.',9,41),
 ('togoria','high_ridge_hunting_ground','Охотничьи склоны высокой гряды','Сухой наземный путь огибает выступы над широкой равниной. Крупные звери и липхоны оспаривают укрытия у этой тропы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Eres III.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('eres_iii','survey_supply_landing','Посадочная площадка исследователей','Небольшой настил стоит возле временных складов экспедиции. Грабители заняли пустые навесы рядом с брошенными приборами.',1,1),
 ('eres_iii','weathered_plain_track','Тропа выветренной равнины','Сухая дорога проходит между камнями и пятнами низкой травы. Звери занимают норы и углубления по краям пути.',2,6),
 ('eres_iii','old_scar_observation_post','Пост у старого разлома','Лёгкие исследовательские укрытия стоят возле следов давних разрушений. Повреждённые машины продолжают обходы, а грабители ищут уцелевшее снаряжение.',3,11),
 ('eres_iii','xoxin_outer_grasslands','Внешние участки равнин Ксоксин','Полосы травы чередуются с обнажённым грунтом и старыми следами огня. В укрытых ложбинах живут роющие звери и местные хищники.',4,16),
 ('eres_iii','abandoned_sample_sheds','Покинутые навесы образцов','На пустой площадке сохранились контейнеры и переносные приборы. Сбойные исследовательские дроиды охраняют свои старые рабочие места.',5,21),
 ('eres_iii','raider_supply_hollows','Ложбины тайных складов','Грабители спрятали похищенное снаряжение за складками местности. Они удерживают подходы к навесам и переделанной технике.',6,26),
 ('eres_iii','regrowing_stone_valley','Зарастающая каменная долина','За грядой сохранились защищённые участки травы и кустарника. Звери используют глубокие норы среди следов старого разорения.',7,31),
 ('eres_iii','deep_burrow_banks','Откосы глубоких нор','Сухие склоны пронизаны ходами и короткими боковыми лазами. Крупные роющие звери защищают выходы к наземной тропе.',8,36),
 ('eres_iii','remote_survey_compound','Дальний исследовательский двор','Полевые модули окружают склад привозного оборудования. Грабители заняли помещения, оставив повреждённые машины на подходах.',9,41),
 ('eres_iii','xoxin_scarred_ridges','Изрезанные гряды Ксоксин','Каменный маршрут идёт между старыми шрамами равнин и заросшими ложбинами. Здесь скрываются крупные хищники и звери из глубоких нор.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Taanab.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('taanab','jungle_survey_landing','Разведочная площадка в джунглях','Небольшой расчищенный участок окружён густым лесом. Грабители заняли временные навесы экспедиции рядом с оставленными машинами.',1,1),
 ('taanab','dense_root_trail','Тропа густых корней','Узкий проход идёт между переплетёнными корнями и лианами. Ползуны занимают боковые полости, а хищники наблюдают из подлеска.',2,6),
 ('taanab','abandoned_field_tents','Покинутые полевые палатки','От маленького лагеря остались лёгкие укрытия и ящики приборов. Налётчики используют их, избегая сбойных разведочных машин.',3,11),
 ('taanab','creek_bank_passage','Проход вдоль лесного ручья','Сухая береговая тропа огибает густые заросли у воды. Звери прячутся под корнями и у каменных перекатов.',4,16),
 ('taanab','fallen_trunk_hollows','Полости поваленных стволов','Старые деревья образовали укрытые переходы в подлеске. В корневых нишах поселились ползуны, а рядом проходят охотничьи тропы.',5,21),
 ('taanab','raider_supply_clearing','Поляна припасов налётчиков','Грабители огородили небольшую лесную площадку ящиками. Они охраняют украденное снаряжение и переделанную технику приезжих.',6,26),
 ('taanab','vine_shrouded_ridge','Гряда под покровом лиан','Каменная тропа поднимается над влажными низинами, не покидая лесного полога. Хищники занимают скальные ниши и заросшие уступы.',7,31),
 ('taanab','forgotten_sensor_huts','Забытые хижины наблюдателей','Переносные приборные посты почти скрылись в зелени. Повреждённые разведочные машины продолжают обходить тесные площадки.',8,36),
 ('taanab','deep_root_den_paths','Тропы глубоких корневых логовищ','Крупные корни образуют сухие полости под плотным подлеском. Звери охраняют узкие выходы к лесной дорожке.',9,41),
 ('taanab','remote_jungle_hunting_ground','Дальние охотничьи тропы джунглей','За каменной грядой лес становится особенно густым. Крупные хищники и ползуны занимают укрытия у единственного удобного наземного прохода.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Zeltros.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('zeltros','service_freight_landing','Служебная грузовая площадка','Посадочный настил выходит к закрытым хозяйственным складам. Похитители грузов заняли пустые ячейки вдали от улиц отдыха.',1,1),
 ('zeltros','backstage_supply_lane','Переулок за павильонами','Тесный служебный проезд соединяет складские двери и погрузочные навесы. Преступники используют его для перевозки краденых товаров.',2,6),
 ('zeltros','closed_scenery_workshop','Закрытая мастерская декораций','Под высокими перекрытиями остались рамы и монтажное оборудование. Сбойные машины продолжают работу среди пустых стендов.',3,11),
 ('zeltros','abandoned_garden_service_yard','Покинутый двор садовой службы','За оградой стоят хозяйственные будки и старые механизмы. Повреждённые дроиды обходят закрытый участок обслуживания.',4,16),
 ('zeltros','seized_delivery_storehouses','Захваченные склады доставки','Банда разместила похищенные грузы в пустующих служебных корпусах. Вооружённые посты охраняют подходы к внутренним дворам.',5,21),
 ('zeltros','underground_service_corridors','Подземные служебные коридоры','Технические переходы соединяют закрытые павильоны со складами. Налётчики перемещают здесь добычу рядом со сбойной хозяйственной техникой.',6,26),
 ('zeltros','idle_pavilion_machine_halls','Машинные залы пустого павильона','Служебные помещения давно закрытого здания сохранили оборудование оформления. Повреждённые дроиды продолжают двигаться между монтажными площадками.',7,31),
 ('zeltros','outlaw_workshop_courts','Дворы подпольных мастерских','Высокие стены скрывают ремонтные навесы преступной группы. Боевики охраняют запасы и переделанные монтажные машины.',8,36),
 ('zeltros','hidden_freight_vaults','Тайные грузовые хранилища','Склады банды расположены за закрытыми внутренними проездами. Похитители и вооружённая охрана перекрывают проходы между контейнерами.',9,41),
 ('zeltros','fortified_service_pavilion','Укреплённый служебный павильон','Заброшенный хозяйственный корпус стал опорным пунктом преступников. Боевики и сбойная техника охраняют внутренние склады.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Cerea.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('cerea','outland_supply_landing','Окраинная площадка приезжих','Малый пункт снабжения стоит у начала полевого маршрута. Грабители заняли пустующие навесы с привозным оборудованием.',1,1),
 ('cerea','grassland_edge_path','Тропа края равнины','Сухая дорожка огибает заросшие ложбины за пределами хозяйственных участков. В кустарнике скрываются местные хищники.',2,6),
 ('cerea','old_river_loading_post','Старый речной погрузочный пост','У берега сохранились лёгкие грузовые навесы приезжих. Налётчики используют их рядом с брошенной рабочей машиной.',3,11),
 ('cerea','palm_grove_trail','Тропа пальмовой рощи','Пальмовидные деревья затеняют путь между равниной и рекой. Хищники занимают густые заросли вдоль сухого прохода.',4,16),
 ('cerea','quiet_riverbank_hollows','Ложбины тихого берега','Наземная тропа проходит над водой среди кустарника и корней. Речные звери защищают свои укрытия у береговых выступов.',5,21),
 ('cerea','raider_field_store','Полевой склад грабителей','Приезжая банда устроила тайник в покинутых хозяйственных навесах. Она охраняет украденные припасы и захваченные машины.',6,26),
 ('cerea','sheltered_plain_gullies','Укрытые овраги равнины','Пологие склоны закрывают густые участки кустарника от ветра. Крупные хищники используют боковые ниши как логовища.',7,31),
 ('cerea','abandoned_outland_workyard','Покинутый рабочий двор приезжих','Привозное оборудование осталось на небольшой площадке у дороги. Сбойные машины движутся среди пустых ящиков и ремонтных навесов.',8,36),
 ('cerea','deep_river_thicket_path','Тропа глубоких речных зарослей','Сухие гряды соединяют плотные кустарниковые участки вдоль реки. Хищники занимают укрытия у узких выходов к воде.',9,41),
 ('cerea','remote_outlander_hideout','Дальнее убежище грабителей','Небольшая приезжая группа укрепила старый пункт снабжения за холмами. Налётчики и переделанная техника охраняют запасы похищенных грузов.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Bimmisaari.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('bimmisaari','forest_expedition_landing','Посадочная поляна экспедиции','Небольшой расчищенный участок окружён лесом асаари. Браконьеры заняли пустые палаточные площадки рядом с оставленной техникой приезжих.',1,1),
 ('bimmisaari','asaari_root_paths','Корневые тропы асаари','Наземный путь проходит между густым подлеском и крупными стволами. Хищники и ползуны занимают полости вдоль переплетённых корней.',2,6),
 ('bimmisaari','abandoned_survey_shelters','Покинутые укрытия наблюдателей','Лёгкие навесы сохранились у небольшой лесной поляны. Браконьеры используют ящики экспедиции и повреждённые рабочие машины.',3,11),
 ('bimmisaari','plain_woodland_boundary','Граница равнины и леса','Травянистый участок постепенно переходит в плотную зелень. Хищники подстерегают добычу возле тенистых выходов из подлеска.',4,16),
 ('bimmisaari','fallen_asaari_hollows','Полости поваленных асаари','Старые стволы образуют укрытия на влажной лесной почве. Ползуны защищают входы в тёмные пространства между корнями.',5,21),
 ('bimmisaari','poacher_hidden_clearing','Тайная поляна браконьеров','Приезжие преступники спрятали снаряжение за плотной стеной зелени. Они охраняют временные навесы и переделанные машины.',6,26),
 ('bimmisaari','shaded_jungle_stream','Тропа тенистого лесного ручья','Сухой береговой путь огибает камни и заросшие углубления. Звери скрываются под корнями у тихой воды.',7,31),
 ('bimmisaari','forgotten_instrument_camp','Забытый приборный лагерь','Переносные посты наблюдения почти скрылись в подлеске. Повреждённые дроиды продолжают обходить маленькие рабочие площадки.',8,36),
 ('bimmisaari','deep_forest_den_trail','Тропа глубоких лесных логовищ','Наземный проход тянется между плотными корневыми укрытиями. Крупные хищники и ползуны охраняют выходы на лесные поляны.',9,41),
 ('bimmisaari','remote_poacher_stockade','Дальняя стоянка браконьеров','Самодельная ограда скрывает временные склады преступной группы. Приезжие охраняют похищенное снаряжение и машины у единственного входа.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('dorin','sealed_arrival_hall','dorin_cargo_raider','dorin_service_droid','dorin_cargo_raider'),
 ('dorin','inner_transfer_passage','dorin_cargo_raider','dorin_service_droid','dorin_outlaw_technician'),
 ('dorin','closed_instrument_store','dorin_outlaw_technician','dorin_survey_droid','dorin_outlaw_technician'),
 ('dorin','protected_observation_gallery','dorin_survey_droid','dorin_survey_droid','dorin_service_droid'),
 ('dorin','sealed_maintenance_annex','dorin_cargo_raider','dorin_service_droid','dorin_outlaw_technician'),
 ('dorin','isolated_sensor_laboratory','dorin_survey_droid','dorin_survey_droid','dorin_survey_droid'),
 ('dorin','covered_cargo_locks','dorin_cargo_raider','dorin_service_droid','dorin_outlaw_technician'),
 ('dorin','outlaw_pressure_workshops','dorin_outlaw_technician','dorin_service_droid','dorin_outlaw_technician'),
 ('dorin','deep_service_sections','dorin_service_droid','dorin_survey_droid','dorin_service_droid'),
 ('dorin','fortified_guest_depot','dorin_cargo_raider','dorin_outlaw_technician','dorin_service_droid'),
 ('glee_anselm','sheltered_island_landing','glee_anselm_cove_raider','glee_anselm_pier_droid','glee_anselm_cove_raider'),
 ('glee_anselm','coastal_thicket_path','glee_anselm_island_stalker','glee_anselm_tidal_crawler','glee_anselm_island_stalker'),
 ('glee_anselm','abandoned_fishing_piers','glee_anselm_cove_raider','glee_anselm_pier_droid','glee_anselm_pier_droid'),
 ('glee_anselm','tidal_pool_walkway','glee_anselm_tidal_crawler','glee_anselm_tidal_crawler','glee_anselm_tidal_crawler'),
 ('glee_anselm','wooded_island_ridge','glee_anselm_island_stalker','glee_anselm_island_stalker','glee_anselm_island_stalker'),
 ('glee_anselm','raider_boathouse_yards','glee_anselm_cove_raider','glee_anselm_pier_droid','glee_anselm_cove_raider'),
 ('glee_anselm','wetland_boardwalk','glee_anselm_tidal_crawler','glee_anselm_island_stalker','glee_anselm_island_stalker'),
 ('glee_anselm','stone_shore_den_path','glee_anselm_tidal_crawler','glee_anselm_tidal_crawler','glee_anselm_island_stalker'),
 ('glee_anselm','hidden_cove_storehouses','glee_anselm_cove_raider','glee_anselm_pier_droid','glee_anselm_cove_raider'),
 ('glee_anselm','fortified_island_pier','glee_anselm_cove_raider','glee_anselm_pier_droid','glee_anselm_pier_droid'),
 ('gand','expedition_arrival_lock','gand_expedition_raider','gand_freight_droid','gand_expedition_raider'),
 ('gand','inner_sample_passage','gand_sample_droid','gand_freight_droid','gand_sample_droid'),
 ('gand','sealed_supply_rooms','gand_expedition_raider','gand_smuggler_guard','gand_smuggler_guard'),
 ('gand','protected_mist_observatory','gand_sample_droid','gand_sample_droid','gand_sample_droid'),
 ('gand','closed_loading_chambers','gand_freight_droid','gand_freight_droid','gand_freight_droid'),
 ('gand','seized_expedition_store','gand_expedition_raider','gand_smuggler_guard','gand_freight_droid'),
 ('gand','isolated_instrument_wing','gand_sample_droid','gand_sample_droid','gand_freight_droid'),
 ('gand','sealed_return_galleries','gand_expedition_raider','gand_freight_droid','gand_smuggler_guard'),
 ('gand','outlaw_inner_workrooms','gand_smuggler_guard','gand_sample_droid','gand_smuggler_guard'),
 ('gand','fortified_expedition_vault','gand_smuggler_guard','gand_freight_droid','gand_sample_droid'),
 ('togoria','field_supply_landing','togoria_poacher','togoria_survey_droid','togoria_poacher'),
 ('togoria','tall_grass_track','togoria_plains_stalker','togoria_liphon','togoria_plains_stalker'),
 ('togoria','woodland_edge_paths','togoria_plains_stalker','togoria_plains_stalker','togoria_plains_stalker'),
 ('togoria','abandoned_field_sheds','togoria_poacher','togoria_survey_droid','togoria_poacher'),
 ('togoria','windy_hill_pass','togoria_plains_stalker','togoria_liphon','togoria_liphon'),
 ('togoria','poacher_hidden_camp','togoria_poacher','togoria_survey_droid','togoria_poacher'),
 ('togoria','deep_thicket_hollows','togoria_plains_stalker','togoria_plains_stalker','togoria_plains_stalker'),
 ('togoria','liphon_rock_ledges','togoria_liphon','togoria_liphon','togoria_liphon'),
 ('togoria','remote_poacher_store','togoria_poacher','togoria_survey_droid','togoria_poacher'),
 ('togoria','high_ridge_hunting_ground','togoria_plains_stalker','togoria_liphon','togoria_liphon'),
 ('eres_iii','survey_supply_landing','eres_iii_salvage_raider','eres_iii_survey_droid','eres_iii_salvage_raider'),
 ('eres_iii','weathered_plain_track','eres_iii_burrower','eres_iii_plain_stalker','eres_iii_burrower'),
 ('eres_iii','old_scar_observation_post','eres_iii_salvage_raider','eres_iii_survey_droid','eres_iii_survey_droid'),
 ('eres_iii','xoxin_outer_grasslands','eres_iii_burrower','eres_iii_plain_stalker','eres_iii_plain_stalker'),
 ('eres_iii','abandoned_sample_sheds','eres_iii_survey_droid','eres_iii_survey_droid','eres_iii_survey_droid'),
 ('eres_iii','raider_supply_hollows','eres_iii_salvage_raider','eres_iii_survey_droid','eres_iii_salvage_raider'),
 ('eres_iii','regrowing_stone_valley','eres_iii_burrower','eres_iii_plain_stalker','eres_iii_plain_stalker'),
 ('eres_iii','deep_burrow_banks','eres_iii_burrower','eres_iii_burrower','eres_iii_burrower'),
 ('eres_iii','remote_survey_compound','eres_iii_salvage_raider','eres_iii_survey_droid','eres_iii_salvage_raider'),
 ('eres_iii','xoxin_scarred_ridges','eres_iii_plain_stalker','eres_iii_burrower','eres_iii_plain_stalker'),
 ('taanab','jungle_survey_landing','taanab_expedition_raider','taanab_survey_droid','taanab_expedition_raider'),
 ('taanab','dense_root_trail','taanab_root_crawler','taanab_jungle_stalker','taanab_root_crawler'),
 ('taanab','abandoned_field_tents','taanab_expedition_raider','taanab_survey_droid','taanab_survey_droid'),
 ('taanab','creek_bank_passage','taanab_root_crawler','taanab_jungle_stalker','taanab_jungle_stalker'),
 ('taanab','fallen_trunk_hollows','taanab_root_crawler','taanab_root_crawler','taanab_jungle_stalker'),
 ('taanab','raider_supply_clearing','taanab_expedition_raider','taanab_survey_droid','taanab_expedition_raider'),
 ('taanab','vine_shrouded_ridge','taanab_jungle_stalker','taanab_root_crawler','taanab_jungle_stalker'),
 ('taanab','forgotten_sensor_huts','taanab_survey_droid','taanab_survey_droid','taanab_survey_droid'),
 ('taanab','deep_root_den_paths','taanab_jungle_stalker','taanab_root_crawler','taanab_root_crawler'),
 ('taanab','remote_jungle_hunting_ground','taanab_jungle_stalker','taanab_root_crawler','taanab_jungle_stalker'),
 ('zeltros','service_freight_landing','zeltros_cargo_thief','zeltros_service_droid','zeltros_cargo_thief'),
 ('zeltros','backstage_supply_lane','zeltros_cargo_thief','zeltros_gang_enforcer','zeltros_cargo_thief'),
 ('zeltros','closed_scenery_workshop','zeltros_stage_droid','zeltros_stage_droid','zeltros_stage_droid'),
 ('zeltros','abandoned_garden_service_yard','zeltros_service_droid','zeltros_service_droid','zeltros_service_droid'),
 ('zeltros','seized_delivery_storehouses','zeltros_cargo_thief','zeltros_gang_enforcer','zeltros_gang_enforcer'),
 ('zeltros','underground_service_corridors','zeltros_cargo_thief','zeltros_service_droid','zeltros_gang_enforcer'),
 ('zeltros','idle_pavilion_machine_halls','zeltros_stage_droid','zeltros_service_droid','zeltros_stage_droid'),
 ('zeltros','outlaw_workshop_courts','zeltros_gang_enforcer','zeltros_stage_droid','zeltros_gang_enforcer'),
 ('zeltros','hidden_freight_vaults','zeltros_cargo_thief','zeltros_gang_enforcer','zeltros_gang_enforcer'),
 ('zeltros','fortified_service_pavilion','zeltros_gang_enforcer','zeltros_stage_droid','zeltros_service_droid'),
 ('cerea','outland_supply_landing','cerea_outlander_raider','cerea_field_droid','cerea_outlander_raider'),
 ('cerea','grassland_edge_path','cerea_grassland_stalker','cerea_grassland_stalker','cerea_grassland_stalker'),
 ('cerea','old_river_loading_post','cerea_outlander_raider','cerea_field_droid','cerea_outlander_raider'),
 ('cerea','palm_grove_trail','cerea_grassland_stalker','cerea_riverbank_predator','cerea_grassland_stalker'),
 ('cerea','quiet_riverbank_hollows','cerea_riverbank_predator','cerea_riverbank_predator','cerea_riverbank_predator'),
 ('cerea','raider_field_store','cerea_outlander_raider','cerea_field_droid','cerea_outlander_raider'),
 ('cerea','sheltered_plain_gullies','cerea_grassland_stalker','cerea_grassland_stalker','cerea_grassland_stalker'),
 ('cerea','abandoned_outland_workyard','cerea_field_droid','cerea_field_droid','cerea_field_droid'),
 ('cerea','deep_river_thicket_path','cerea_riverbank_predator','cerea_grassland_stalker','cerea_riverbank_predator'),
 ('cerea','remote_outlander_hideout','cerea_outlander_raider','cerea_field_droid','cerea_outlander_raider'),
 ('bimmisaari','forest_expedition_landing','bimmisaari_offworld_poacher','bimmisaari_survey_droid','bimmisaari_offworld_poacher'),
 ('bimmisaari','asaari_root_paths','bimmisaari_thicket_crawler','bimmisaari_forest_stalker','bimmisaari_thicket_crawler'),
 ('bimmisaari','abandoned_survey_shelters','bimmisaari_offworld_poacher','bimmisaari_survey_droid','bimmisaari_survey_droid'),
 ('bimmisaari','plain_woodland_boundary','bimmisaari_forest_stalker','bimmisaari_forest_stalker','bimmisaari_forest_stalker'),
 ('bimmisaari','fallen_asaari_hollows','bimmisaari_thicket_crawler','bimmisaari_thicket_crawler','bimmisaari_thicket_crawler'),
 ('bimmisaari','poacher_hidden_clearing','bimmisaari_offworld_poacher','bimmisaari_survey_droid','bimmisaari_offworld_poacher'),
 ('bimmisaari','shaded_jungle_stream','bimmisaari_thicket_crawler','bimmisaari_forest_stalker','bimmisaari_forest_stalker'),
 ('bimmisaari','forgotten_instrument_camp','bimmisaari_survey_droid','bimmisaari_survey_droid','bimmisaari_survey_droid'),
 ('bimmisaari','deep_forest_den_trail','bimmisaari_forest_stalker','bimmisaari_thicket_crawler','bimmisaari_forest_stalker'),
 ('bimmisaari','remote_poacher_stockade','bimmisaari_offworld_poacher','bimmisaari_survey_droid','bimmisaari_offworld_poacher')
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
WHERE p.code IN ('dorin','glee_anselm','gand','togoria','eres_iii','taanab','zeltros','cerea','bimmisaari') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
