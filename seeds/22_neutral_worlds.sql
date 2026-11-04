-- Sixth neutral-world batch completes content for all 100 catalog planets.
-- Early Great Galactic War: authored districts and roles, informed by Legends.
-- Kessel's early sealed ground/underground complex is a game adaptation.
-- Eriadu precedes Quintad industrialization; Bestine IV has only small field camps.
-- Sluis Van/Fondor locations are ground facilities, not orbital shipyards.
-- Mygeeto's icy appearance is an era adaptation; no late banking cities or characters.
-- No ownership changes, peaceful native enemy categories or new hazard mechanics.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('malastare_woodland_stalker','Хищник лесов Маластара','Прячется в густом подлеске и тенистых ложбинах. Выслеживает добычу вдоль троп, ведущих к открытым каменным участкам.','beast'),
 ('malastare_crag_hunter','Охотник сухих уступов','Обитает среди каменных гряд и сухих расщелин Маластара. Защищает укрытия на склонах над низинами.','beast'),
 ('malastare_fuel_raider','Налётчик топливных складов','Грабит перевозчиков и занимает покинутые хозяйственные посты. Охраняет запасы краденого оборудования и припасов.','humanoid'),
 ('malastare_pump_droid','Сбойный дроид перекачки','Остался у остановленного топливного поста. Повреждённые датчики заставляют его применять рабочие захваты против посетителей.','droid'),
 ('kessel_cargo_raider','Грабитель закрытых складов Кесселя','Занял часть герметичного горного комплекса. Похищает приборы и грузы, оставаясь в помещениях с поддерживаемой атмосферой.','humanoid'),
 ('kessel_mining_droid','Сбойный проходческий дроид','Продолжает движение внутри закрытых выработок. Повреждённое управление заставляет его применять рабочие инструменты против посетителей.','droid'),
 ('kessel_survey_droid','Повреждённый геологический дроид','Обследовал грунт в герметичных исследовательских секциях. Ошибочные команды заставляют машину атаковать движущиеся объекты.','droid'),
 ('kessel_outlaw_guard','Охранник подземных тайников','Вооружённый участник преступной группы удерживает внутренние склады. Защищает украденное оборудование в закрытых коридорах комплекса.','humanoid'),
 ('eriadu_shore_stalker','Хищник береговых расщелин','Живёт среди каменных уступов Эриаду. Подстерегает добычу на сухих береговых тропах и защищает тенистые логовища.','beast'),
 ('eriadu_scrub_predator','Хищник прибрежных зарослей','Скрывается в густом кустарнике на природных окраинах. Выслеживает добычу возле ложбин и небольших речных проток.','beast'),
 ('eriadu_cargo_raider','Грабитель ремесленных грузов','Отнимает припасы и товары у перевозчиков. Прячет добычу в покинутых береговых навесах и небольших мастерских.','humanoid'),
 ('eriadu_craft_droid','Повреждённый ремесленный дроид','Остался у закрытого рабочего двора. Сбойное распознавание заставляет машину применять инструменты против посетителей.','droid'),
 ('sluis_van_cargo_raider','Налётчик наземных складов','Перехватывает грузы на служебных площадках Слуис-Вана. Прячет похищенные детали в закрытых секциях снабжения.','humanoid'),
 ('sluis_van_assembly_droid','Сбойный сборочный дроид','Продолжает выполнять старые команды в наземной мастерской. Повреждённое распознавание заставляет его хватать посетителей вместо деталей.','droid'),
 ('sluis_van_repair_droid','Повреждённый ремонтный дроид','Остался в остановленном служебном корпусе. Применяет рабочие инструменты против движущихся объектов, считая их неисправным оборудованием.','droid'),
 ('sluis_van_outlaw_guard','Охранник складской банды','Вооружённый преступник удерживает захваченные дворы снабжения. Защищает хранилища краденых деталей и подпольные мастерские.','humanoid'),
 ('fondor_welding_droid','Сбойный сварочный дроид','Работал на наземной производственной линии Фондора. Повреждённое управление заставляет его применять инструменты против посетителей.','droid'),
 ('fondor_extraction_droid','Повреждённый проходческий дроид','Остался в старых служебных выработках. Продолжает движение по рабочим маршрутам и атакует любые помехи.','droid'),
 ('fondor_cargo_raider','Похититель заводских грузов','Перехватывает партии деталей и оборудования. Хранит добычу в закрытых заводских корпусах и подземных складах.','humanoid'),
 ('fondor_gang_enforcer','Боевик промышленной банды','Охраняет захваченные производственные дворы. Удерживает входы в подпольные мастерские и тайники с краденым.','humanoid'),
 ('bestine_iv_shore_crab','Краб островных мелководий','Обитает возле каменных берегов Бестина IV и способен выходить на сушу. Защищает укрытия в щелях и под временными настилами.','beast'),
 ('bestine_iv_cliff_hunter','Хищник островных утёсов','Скрывается среди сухих скальных выступов. Выслеживает добычу на узких тропах над морем.','beast'),
 ('bestine_iv_expedition_raider','Грабитель островных стоянок','Приезжий преступник нападает на небольшие экспедиционные посты. Прячет припасы в пустующих навесах у берега.','humanoid'),
 ('bestine_iv_survey_droid','Сбойный островной разведчик','Остался на временном приборном посту. Повреждённые датчики заставляют машину нападать на движущиеся объекты.','droid'),
 ('denon_cargo_thief','Похититель грузов Денона','Вскрывает контейнеры в закрытых грузовых кварталах. Прячет добычу в служебных дворах и пустующих кладовых.','humanoid'),
 ('denon_gang_enforcer','Боевик нижних складов','Вооружённый преступник удерживает захваченные хозяйственные секции. Охраняет тайники банды и проходы к подпольным мастерским.','humanoid'),
 ('denon_loader_droid','Сбойный городской погрузчик','Продолжает работать на остановленной перевалочной линии. Повреждённое распознавание заставляет его хватать посетителей вместо грузов.','droid'),
 ('denon_utility_droid','Повреждённый коммунальный дроид','Остался в закрытых технических коридорах. Ошибочные команды заставляют машину атаковать движущиеся объекты рабочими инструментами.','droid'),
 ('mygeeto_ice_stalker','Хищник ледяных расщелин','Прячется среди снежных уступов и кристальных гряд Майгито. Охотится возле защищённых от ветра проходов.','beast'),
 ('mygeeto_ice_worm','Ледяной червь Майгито','Крупное червеобразное существо укрывается в толще ледяных склонов. Выходит к поверхности возле разломов и защищает своё логовище.','beast'),
 ('mygeeto_crystal_raider','Грабитель горных грузов','Перехватывает оборудование и минеральные грузы на удалённых постах. Держит тайники в покинутых служебных помещениях.','humanoid'),
 ('mygeeto_tunnel_droid','Повреждённый горный дроид','Остался у старых проходческих площадок. Повреждённая программа заставляет его нападать на посетителей рабочими инструментами.','droid'),
 ('muunilinst_plain_stalker','Хищник равнин Муунилинста','Охотится в травянистых ложбинах и на лесных окраинах. Подстерегает добычу возле тенистых переходов между холмами.','beast'),
 ('muunilinst_mountain_hunter','Хищник горных склонов','Живёт среди каменных гряд и лесистых предгорий Муунилинста. Защищает логовища в сухих скальных нишах.','beast'),
 ('muunilinst_metal_raider','Грабитель минеральных грузов','Перехватывает партии металла и припасы перевозчиков. Прячет добычу в покинутых перевалочных дворах.','humanoid'),
 ('muunilinst_ore_droid','Сбойный дроид обработки руды','Остался у закрытого рабочего поста. Повреждённое управление заставляет его атаковать посетителей инструментами и захватами.','droid')
ON CONFLICT (code) DO NOTHING;

-- Malastare.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('malastare','outer_supply_landing','Внешняя площадка снабжения','Посадочный настил примыкает к старым грузовым навесам. Налётчики заняли пустующие склады рядом с брошенными рабочими машинами.',1,1),
 ('malastare','forest_edge_track','Тропа лесной окраины','Сухой путь идёт между высокой зеленью и каменными выступами. Хищники наблюдают из подлеска и тенистых расщелин.',2,6),
 ('malastare','idle_transfer_station','Остановленный перевалочный пост','В ограждённом дворе сохранились закрытые ёмкости и навесы. Сбойные машины продолжают обходы среди тайников налётчиков.',3,11),
 ('malastare','dry_rock_gullies','Овраги сухих скал','Каменные ложбины разделяют засушливую часть равнины. Местные звери занимают ниши, защищённые от солнца и ветра.',4,16),
 ('malastare','wooded_river_bluffs','Лесистые речные уступы','Тропа проходит по прочному береговому краю под густыми кронами. Хищники держатся у заросших выходов к каменным полкам.',5,21),
 ('malastare','raider_storage_yard','Складской двор налётчиков','Банда укрепила покинутый пункт снабжения контейнерами. Вооружённые посты и переделанные машины охраняют украденные грузы.',6,26),
 ('malastare','methane_basin_overlook','Уступ над метановыми низинами','Сухой каменный маршрут проходит высоко над озёрными впадинами. Звери занимают боковые расщелины вдоль защищённой тропы.',7,31),
 ('malastare','closed_service_workshops','Закрытые служебные мастерские','Старые ремонтные корпуса отделены от резервуаров глухими стенами. Налётчики разбирают здесь оборудование рядом со сбойными машинами.',8,36),
 ('malastare','deep_forest_hunting_paths','Охотничьи тропы глубокого леса','Плотный подлесок скрывает сухие ложбины и корневые укрытия. Крупные хищники защищают выходы к редким открытым полянам.',9,41),
 ('malastare','fortified_supply_compound','Укреплённый пункт снабжения','В заброшенном хозяйственном комплексе банда разместила основные запасы. Налётчики и повреждённая техника удерживают внутренние проезды.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Kessel.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('kessel','sealed_mining_arrival','Герметичный горный ангар','Закрытая посадочная секция соединена шлюзами с подземным комплексом. Внутри поддерживается пригодный воздух, а пустующие грузовые ячейки заняли грабители.',1,1),
 ('kessel','covered_transfer_galleries','Крытые перевалочные галереи','Герметичные переходы связывают ангар со служебными складами. Сбойные машины обходят контейнеры, которые преступники используют для укрытия.',2,6),
 ('kessel','isolated_geology_rooms','Изолированные геологические камеры','Исследовательские помещения отделены от разреженной внешней атмосферы. Повреждённые дроиды продолжают обследовать пустые рабочие площадки.',3,11),
 ('kessel','sealed_boring_passage','Герметичный проход выработки','Закреплённый тоннель оборудован шлюзами и исправной подачей воздуха. Неуправляемые проходческие машины движутся вдоль старых рабочих полос.',4,16),
 ('kessel','seized_underground_store','Захваченный подземный склад','Преступники укрепили кладовые внутри защитного контура комплекса. Они охраняют похищенные приборы, сохраняя работающую систему воздуха.',5,21),
 ('kessel','inner_service_workshops','Внутренние служебные мастерские','Закрытые ремонтные помещения соединены безопасными переходами. Сбойные машины работают среди запасов оборудования, присвоенных бандой.',6,26),
 ('kessel','pressurized_deep_gallery','Глубокая герметичная галерея','Тоннель уходит под каменные пласты, не покидая контура жизнеобеспечения. Повреждённые дроиды обходят пустые места исследований.',7,31),
 ('kessel','outlaw_inner_loading_halls','Внутренние погрузочные залы банды','Преступники перекрыли переходы между закрытыми грузовыми камерами. Вооружённая охрана и проходческие машины защищают тайники.',8,36),
 ('kessel','sealed_instrument_vaults','Герметичные хранилища приборов','Кладовые скрыты за внутренними техническими дверями. Грабители удерживают запасы в секциях с исправной подачей воздуха.',9,41),
 ('kessel','fortified_underground_depot','Укреплённое подземное хранилище','Основной склад банды находится внутри защищённого корпуса. Охрана и переделанные машины удерживают подходы из герметичных галерей.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Eriadu.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('eriadu','coastal_supply_landing','Береговая площадка снабжения','Небольшой посадочный настил примыкает к ремесленным складам. Грабители заняли пустые навесы рядом с оставленной рабочей машиной.',1,1),
 ('eriadu','rocky_shore_path','Тропа каменного берега','Сухой путь огибает прибрежные выступы и тенистые ниши. Местные хищники занимают расщелины у воды.',2,6),
 ('eriadu','idle_shellwork_court','Пустующий двор обработки раковин','В закрытых навесах сохранились столы, инструменты и упаковочные ящики. Грабители сортируют здесь добычу рядом со сбойной ремесленной техникой.',3,11),
 ('eriadu','coastal_scrub_hollows','Ложбины береговых зарослей','Густой кустарник покрывает защищённые участки между каменными грядами. Звери наблюдают за тропой из тенистых укрытий.',4,16),
 ('eriadu','old_river_transfer_post','Старый речной перевалочный пост','Лёгкие грузовые площадки стоят у спокойного берегового участка. Банда заняла пустые кладовые с товарами и небольшими рабочими машинами.',5,21),
 ('eriadu','outlaw_craft_workshops','Ремесленные мастерские грабителей','Похищенные ящики сложены в небольших закрытых дворах. Преступники охраняют запасы и переделывают уцелевшее оборудование.',6,26),
 ('eriadu','inland_thicket_ridge','Гряда внутренних зарослей','Сухая тропа поднимается от берега к кустарниковым ложбинам. Крупные хищники занимают боковые ниши среди камней.',7,31),
 ('eriadu','abandoned_repair_sheds','Заброшенные ремонтные навесы','В стороне от дороги сохранился малый рабочий пост. Повреждённые машины продолжают обходить пустые площадки обслуживания.',8,36),
 ('eriadu','deep_shore_den_trail','Тропа глубоких береговых логовищ','Каменный маршрут связывает укрытые расщелины и густые заросли. Звери защищают выходы к тихим участкам побережья.',9,41),
 ('eriadu','fortified_coastal_store','Укреплённый береговой склад','Банда оградила старый перевалочный двор рядами контейнеров. Налётчики и повреждённые ремесленные машины охраняют похищенные товары.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Sluis Van.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('sluis_van','surface_supply_landing','Наземная площадка снабжения','Посадочный настил выходит к складским корпусам судостроительных поставок. Налётчики заняли пустующие секции с деталями и ремонтным оборудованием.',1,1),
 ('sluis_van','component_transfer_yards','Дворы перегрузки компонентов','Ряды контейнеров разделяют служебную площадку на узкие проезды. Преступники охраняют спрятанные здесь партии краденых деталей.',2,6),
 ('sluis_van','idle_assembly_workshops','Остановленные сборочные мастерские','Наземные цеха сохранили рабочие линии и пустые стенды. Сбойные дроиды продолжают движение между сборочными постами.',3,11),
 ('sluis_van','closed_repair_annex','Закрытый ремонтный корпус','Служебные помещения соединены крытыми переходами. Повреждённые машины обслуживают пустые ремонтные места без операторов.',4,16),
 ('sluis_van','seized_parts_depot','Захваченный склад деталей','Банда удерживает несколько грузовых ангаров на окраине комплекса. Вооружённые посты охраняют контейнеры и украденную рабочую технику.',5,21),
 ('sluis_van','underground_supply_passages','Подземные проходы снабжения','Технические тоннели связывают наземные склады с мастерскими. Налётчики перевозят здесь добычу рядом со сбойными машинами.',6,26),
 ('sluis_van','abandoned_test_halls','Покинутые испытательные залы','Закрытый корпус использовался для проверки небольших сборочных узлов. Неуправляемые дроиды продолжают обходить пустые стенды.',7,31),
 ('sluis_van','outlaw_repair_courts','Ремонтные дворы преступной группы','Преступники оградили старые мастерские грузовыми контейнерами. Охрана удерживает проходы к переделанным машинам и складам.',8,36),
 ('sluis_van','deep_component_store','Дальние хранилища компонентов','Укреплённые кладовые скрыты за служебными корпусами. Налётчики и охранники банды защищают основные запасы похищенных деталей.',9,41),
 ('sluis_van','fortified_surface_terminal','Укреплённый наземный терминал','Старый узел снабжения превращён в опорный пункт банды. Вооружённые посты и переделанные дроиды перекрывают внутренние грузовые маршруты.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Fondor.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('fondor','ground_factory_landing','Наземная заводская площадка','Посадочный настил соединён с пустующими грузовыми корпусами. Похитители деталей заняли старые ячейки рядом с рабочей техникой.',1,1),
 ('fondor','outer_machine_yards','Внешние машинные дворы','Контейнеры и снятые корпуса образуют тесные проезды. Банда использует их как укрытия при перевозке украденного оборудования.',2,6),
 ('fondor','closed_welding_halls','Закрытые сварочные залы','На остановленных линиях остались пустые крепления и монтажные площадки. Сбойные дроиды продолжают выполнять рабочие движения.',3,11),
 ('fondor','cooled_service_galleries','Остывшие служебные галереи','Технические проходы идут вдоль отключённых труб и старых опор. Повреждённые машины обходят пустые места обслуживания.',4,16),
 ('fondor','seized_component_warehouses','Захваченные склады компонентов','Преступники удерживают ряды закрытых ангаров. Боевики охраняют похищенные детали и рабочие машины.',5,21),
 ('fondor','abandoned_boring_sector','Покинутый проходческий сектор','Закреплённые тоннели соединяют остановленные участки выработок. Сбойные дроиды продолжают двигаться по старым грузовым полосам.',6,26),
 ('fondor','outlaw_fabrication_workshops','Подпольные производственные мастерские','Банда приспособила закрытые цеха для переделки украденной техники. Вооружённые посты прикрывают рабочие площадки.',7,31),
 ('fondor','deep_material_galleries','Глубокие галереи материалов','Подземные склады связаны широкими техническими коридорами. Похитители грузов используют их рядом с повреждёнными проходческими машинами.',8,36),
 ('fondor','idle_automated_plant','Остановленный автоматический цех','В закрытом корпусе сохранились линии сборки и подачи материалов. Неуправляемые дроиды продолжают работу без поставок и персонала.',9,41),
 ('fondor','fortified_industrial_depot','Укреплённое промышленное хранилище','Основные запасы банды скрыты за внутренними стенами старого комплекса. Боевики и переделанные машины защищают подходы из служебных галерей.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Bestine IV.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('bestine_iv','island_field_landing','Полевая площадка на острове','Небольшой настил установлен на сухом участке каменного берега. Грабители заняли временные навесы экспедиции рядом с оставленными приборами.',1,1),
 ('bestine_iv','rocky_tidal_path','Тропа каменных приливных берегов','Приподнятая сухая дорожка огибает мелкие заводи. Крабы выходят из воды и защищают щели у скальных опор.',2,6),
 ('bestine_iv','abandoned_survey_shelters','Покинутые укрытия исследователей','Лёгкие постройки стоят на защищённой береговой полке. Налётчики хранят здесь добычу, избегая сбойных разведочных машин.',3,11),
 ('bestine_iv','wind_sheltered_cliff_track','Укрытая тропа утёсов','Скальная стена защищает узкий наземный путь от ветра. Хищники занимают тенистые ниши над берегом.',4,16),
 ('bestine_iv','shallow_cove_walkways','Мостки мелкой бухты','Прочные временные настилы соединяют сухие каменные площадки. Крабы цепляются за опоры и подбираются к краям тропы.',5,21),
 ('bestine_iv','raider_supply_clearing','Стоянка похитителей припасов','Приезжая банда огородила ящиками участок за скальным выступом. Она охраняет украденное снаряжение и захваченную технику.',6,26),
 ('bestine_iv','island_spire_approach','Подступы к островной скале','Сухая тропа поднимается по пологим уступам у высокого каменного пика. Хищники используют боковые расщелины как логовища.',7,31),
 ('bestine_iv','forgotten_tide_instrument_post','Забытый пост наблюдения за приливами','Приборные опоры сохранились на приподнятом настиле над берегом. Повреждённые разведчики продолжают обходить пустые площадки.',8,36),
 ('bestine_iv','deep_coastal_hollows','Глубокие береговые укрытия','Наземный маршрут соединяет сухие ниши и каменные гряды у моря. Крабы и хищники защищают проходы к своим укрытиям.',9,41),
 ('bestine_iv','remote_island_raider_camp','Дальний островной лагерь грабителей','Временные навесы скрыты за кольцом скал на сухой части острова. Налётчики и повреждённые машины удерживают склад экспедиционной добычи.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Denon.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('denon','service_cargo_landing','Служебная грузовая площадка','Наземный посадочный сектор примыкает к старым складам. Похитители контейнеров заняли пустующие ячейки среди повреждённых погрузчиков.',1,1),
 ('denon','lower_storage_lanes','Проезды нижних складов','Служебные дороги проходят под массивными перекрытиями. Преступники укрываются между контейнерами и закрытыми грузовыми дверями.',2,6),
 ('denon','idle_distribution_hall','Остановленный распределительный зал','Старая перевалочная линия сохранила ряды пустых приёмников. Сбойные погрузчики продолжают двигаться по размеченным полосам.',3,11),
 ('denon','closed_utility_passages','Закрытые коммунальные проходы','Технические коридоры соединяют хозяйственные секции квартала. Повреждённые дроиды выполняют устаревшие команды без операторов.',4,16),
 ('denon','seized_freight_courtyards','Захваченные грузовые дворы','Банда укрепила покинутые площадки рядами контейнеров. Боевики охраняют похищенные товары и рабочие машины.',5,21),
 ('denon','underground_transfer_gallery','Подземная перевалочная галерея','Крытый маршрут связывает дальние склады со служебными уровнями. Налётчики перевозят здесь добычу рядом со сбойной коммунальной техникой.',6,26),
 ('denon','outlaw_equipment_workshops','Мастерские похищенного оборудования','Закрытые хозяйственные помещения стали пунктом переделки краденых машин. Вооружённая охрана удерживает подходы к рабочим площадкам.',7,31),
 ('denon','abandoned_service_stack','Покинутый служебный блок','Несколько технических уровней соединены лестницами и внутренними переходами. Сбойные дроиды продолжают обходить пустые помещения.',8,36),
 ('denon','hidden_warehouse_vaults','Тайные складские хранилища','За внутренними дверями спрятаны основные запасы банды. Похитители грузов и вооружённые посты перекрывают узкие проходы.',9,41),
 ('denon','fortified_logistics_block','Укреплённый грузовой блок','Преступная группа заняла дальний корпус старого узла снабжения. Боевики и переделанные дроиды удерживают склады и внутренние проезды.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Mygeeto.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('mygeeto','sheltered_ice_landing','Укрытая ледяная площадка','Небольшой настил защищён скальной стеной от снежных наносов. Грабители заняли пустые грузовые навесы рядом с брошенными машинами.',1,1),
 ('mygeeto','crystal_ridge_path','Тропа кристальной гряды','Каменные полки проходят между ледяными пластами и высокими кристальными выступами. В боковых расщелинах прячутся местные хищники.',2,6),
 ('mygeeto','abandoned_mountain_store','Покинутый горный склад','Служебные корпуса стоят у входов в старые выработки. Налётчики используют кладовые рядом со сбойной проходческой техникой.',3,11),
 ('mygeeto','wind_sheltered_ice_gorge','Укрытое от ветра ледяное ущелье','Извилистый наземный путь проходит по прочному основанию между скалами. Звери занимают снежные ниши вдоль каменных стен.',4,16),
 ('mygeeto','idle_boring_workshops','Остановленные проходческие мастерские','Закрытые навесы соединяют ремонтные площадки с закреплёнными тоннелями. Повреждённые машины продолжают обслуживать пустые рабочие места.',5,21),
 ('mygeeto','raider_crystal_depot','Склад минеральных грузов налётчиков','Преступники укрепили старый горный пост контейнерами. Они охраняют похищенные грузы и уцелевшие проходческие машины.',6,26),
 ('mygeeto','ice_worm_shelf_approach','Подступы к логовищам ледяных червей','Прочный каменный маршрут огибает разломы ледяного склона. Крупные черви выходят из укрытий у поверхности.',7,31),
 ('mygeeto','deep_crystal_service_gallery','Глубокая служебная галерея','Закреплённый тоннель проходит между кристальными пластами. Грабители используют пустые кладовые под защитой повреждённых машин.',8,36),
 ('mygeeto','frozen_worm_hollows','Промёрзшие полости червей','Наземный путь идёт по скальному краю нескольких ледяных разломов. Крупные существа охраняют выходы из своих глубоких логовищ.',9,41),
 ('mygeeto','remote_ice_ridge_dens','Дальние логовища ледяной гряды','Каменные выступы соединяют укрытые от ветра участки склона. Хищники и ледяные черви занимают ниши у самых тесных проходов.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Muunilinst.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('muunilinst','mineral_supply_landing','Площадка минеральных поставок','Небольшой грузовой настил примыкает к старым перевалочным складам. Налётчики заняли пустые навесы рядом с оставленной рабочей техникой.',1,1),
 ('muunilinst','grassy_foothill_track','Тропа травянистых предгорий','Сухой путь идёт между открытыми ложбинами и каменными выступами. Хищники скрываются в высокой траве и тенистых нишах.',2,6),
 ('muunilinst','abandoned_metal_storehouses','Покинутые склады металла','Грузовые дворы сохранили старые ёмкости и закрытые кладовые. Преступники хранят здесь добычу, используя уцелевшие рабочие машины.',3,11),
 ('muunilinst','wooded_valley_trail','Тропа лесистой долины','Густые деревья затеняют проход между предгорьями и равниной. Звери занимают корневые укрытия и выходы к каменным полкам.',4,16),
 ('muunilinst','idle_ore_workyard','Остановленный рудный двор','На небольшом рабочем посту остались пустые приёмники и ремонтные навесы. Сбойные дроиды продолжают выполнять команды без операторов.',5,21),
 ('muunilinst','raider_transfer_compound','Перевалочный пункт грабителей','Банда укрепила старые складские постройки на удалённой дороге. Налётчики охраняют похищенные металлы, припасы и оборудование.',6,26),
 ('muunilinst','jagged_mountain_pass','Проход острых горных гряд','Наземная тропа огибает высокие скальные выступы. Крупные хищники держатся у защищённых ниш и тесных поворотов.',7,31),
 ('muunilinst','sheltered_mineral_coast','Укрытый берег минеральных бухт','Сухой каменный путь проходит вдоль береговых рабочих навесов. Грабители заняли склады у старого оборудования для перегрузки.',8,36),
 ('muunilinst','deep_woodland_hollows','Глубокие ложбины леса','Густой подлесок окружает каменные гряды между равниной и горами. Хищники защищают корневые укрытия и скальные логовища.',9,41),
 ('muunilinst','fortified_metal_depot','Укреплённое хранилище металла','Основные запасы преступной группы скрыты за стенами старого перевалочного корпуса. Налётчики и переделанные машины охраняют внутренние дворы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('malastare','outer_supply_landing','malastare_fuel_raider','malastare_pump_droid','malastare_fuel_raider'),
 ('malastare','forest_edge_track','malastare_woodland_stalker','malastare_crag_hunter','malastare_woodland_stalker'),
 ('malastare','idle_transfer_station','malastare_fuel_raider','malastare_pump_droid','malastare_pump_droid'),
 ('malastare','dry_rock_gullies','malastare_crag_hunter','malastare_crag_hunter','malastare_crag_hunter'),
 ('malastare','wooded_river_bluffs','malastare_woodland_stalker','malastare_crag_hunter','malastare_woodland_stalker'),
 ('malastare','raider_storage_yard','malastare_fuel_raider','malastare_pump_droid','malastare_fuel_raider'),
 ('malastare','methane_basin_overlook','malastare_crag_hunter','malastare_crag_hunter','malastare_crag_hunter'),
 ('malastare','closed_service_workshops','malastare_fuel_raider','malastare_pump_droid','malastare_pump_droid'),
 ('malastare','deep_forest_hunting_paths','malastare_woodland_stalker','malastare_woodland_stalker','malastare_woodland_stalker'),
 ('malastare','fortified_supply_compound','malastare_fuel_raider','malastare_pump_droid','malastare_fuel_raider'),
 ('kessel','sealed_mining_arrival','kessel_cargo_raider','kessel_survey_droid','kessel_cargo_raider'),
 ('kessel','covered_transfer_galleries','kessel_cargo_raider','kessel_mining_droid','kessel_outlaw_guard'),
 ('kessel','isolated_geology_rooms','kessel_survey_droid','kessel_survey_droid','kessel_survey_droid'),
 ('kessel','sealed_boring_passage','kessel_mining_droid','kessel_mining_droid','kessel_mining_droid'),
 ('kessel','seized_underground_store','kessel_cargo_raider','kessel_outlaw_guard','kessel_outlaw_guard'),
 ('kessel','inner_service_workshops','kessel_cargo_raider','kessel_mining_droid','kessel_survey_droid'),
 ('kessel','pressurized_deep_gallery','kessel_survey_droid','kessel_mining_droid','kessel_survey_droid'),
 ('kessel','outlaw_inner_loading_halls','kessel_outlaw_guard','kessel_mining_droid','kessel_outlaw_guard'),
 ('kessel','sealed_instrument_vaults','kessel_cargo_raider','kessel_outlaw_guard','kessel_survey_droid'),
 ('kessel','fortified_underground_depot','kessel_outlaw_guard','kessel_mining_droid','kessel_outlaw_guard'),
 ('eriadu','coastal_supply_landing','eriadu_cargo_raider','eriadu_craft_droid','eriadu_cargo_raider'),
 ('eriadu','rocky_shore_path','eriadu_shore_stalker','eriadu_shore_stalker','eriadu_shore_stalker'),
 ('eriadu','idle_shellwork_court','eriadu_cargo_raider','eriadu_craft_droid','eriadu_craft_droid'),
 ('eriadu','coastal_scrub_hollows','eriadu_scrub_predator','eriadu_shore_stalker','eriadu_scrub_predator'),
 ('eriadu','old_river_transfer_post','eriadu_cargo_raider','eriadu_craft_droid','eriadu_cargo_raider'),
 ('eriadu','outlaw_craft_workshops','eriadu_cargo_raider','eriadu_craft_droid','eriadu_cargo_raider'),
 ('eriadu','inland_thicket_ridge','eriadu_scrub_predator','eriadu_scrub_predator','eriadu_shore_stalker'),
 ('eriadu','abandoned_repair_sheds','eriadu_craft_droid','eriadu_craft_droid','eriadu_craft_droid'),
 ('eriadu','deep_shore_den_trail','eriadu_shore_stalker','eriadu_scrub_predator','eriadu_shore_stalker'),
 ('eriadu','fortified_coastal_store','eriadu_cargo_raider','eriadu_craft_droid','eriadu_cargo_raider'),
 ('sluis_van','surface_supply_landing','sluis_van_cargo_raider','sluis_van_repair_droid','sluis_van_cargo_raider'),
 ('sluis_van','component_transfer_yards','sluis_van_cargo_raider','sluis_van_outlaw_guard','sluis_van_cargo_raider'),
 ('sluis_van','idle_assembly_workshops','sluis_van_assembly_droid','sluis_van_assembly_droid','sluis_van_assembly_droid'),
 ('sluis_van','closed_repair_annex','sluis_van_repair_droid','sluis_van_repair_droid','sluis_van_repair_droid'),
 ('sluis_van','seized_parts_depot','sluis_van_cargo_raider','sluis_van_outlaw_guard','sluis_van_repair_droid'),
 ('sluis_van','underground_supply_passages','sluis_van_cargo_raider','sluis_van_repair_droid','sluis_van_outlaw_guard'),
 ('sluis_van','abandoned_test_halls','sluis_van_assembly_droid','sluis_van_repair_droid','sluis_van_assembly_droid'),
 ('sluis_van','outlaw_repair_courts','sluis_van_outlaw_guard','sluis_van_repair_droid','sluis_van_outlaw_guard'),
 ('sluis_van','deep_component_store','sluis_van_cargo_raider','sluis_van_outlaw_guard','sluis_van_outlaw_guard'),
 ('sluis_van','fortified_surface_terminal','sluis_van_outlaw_guard','sluis_van_assembly_droid','sluis_van_repair_droid'),
 ('fondor','ground_factory_landing','fondor_cargo_raider','fondor_welding_droid','fondor_cargo_raider'),
 ('fondor','outer_machine_yards','fondor_cargo_raider','fondor_gang_enforcer','fondor_cargo_raider'),
 ('fondor','closed_welding_halls','fondor_welding_droid','fondor_welding_droid','fondor_welding_droid'),
 ('fondor','cooled_service_galleries','fondor_extraction_droid','fondor_welding_droid','fondor_extraction_droid'),
 ('fondor','seized_component_warehouses','fondor_cargo_raider','fondor_gang_enforcer','fondor_welding_droid'),
 ('fondor','abandoned_boring_sector','fondor_extraction_droid','fondor_extraction_droid','fondor_extraction_droid'),
 ('fondor','outlaw_fabrication_workshops','fondor_gang_enforcer','fondor_welding_droid','fondor_gang_enforcer'),
 ('fondor','deep_material_galleries','fondor_cargo_raider','fondor_extraction_droid','fondor_gang_enforcer'),
 ('fondor','idle_automated_plant','fondor_welding_droid','fondor_extraction_droid','fondor_welding_droid'),
 ('fondor','fortified_industrial_depot','fondor_gang_enforcer','fondor_welding_droid','fondor_extraction_droid'),
 ('bestine_iv','island_field_landing','bestine_iv_expedition_raider','bestine_iv_survey_droid','bestine_iv_expedition_raider'),
 ('bestine_iv','rocky_tidal_path','bestine_iv_shore_crab','bestine_iv_shore_crab','bestine_iv_shore_crab'),
 ('bestine_iv','abandoned_survey_shelters','bestine_iv_expedition_raider','bestine_iv_survey_droid','bestine_iv_survey_droid'),
 ('bestine_iv','wind_sheltered_cliff_track','bestine_iv_cliff_hunter','bestine_iv_cliff_hunter','bestine_iv_cliff_hunter'),
 ('bestine_iv','shallow_cove_walkways','bestine_iv_shore_crab','bestine_iv_shore_crab','bestine_iv_shore_crab'),
 ('bestine_iv','raider_supply_clearing','bestine_iv_expedition_raider','bestine_iv_survey_droid','bestine_iv_expedition_raider'),
 ('bestine_iv','island_spire_approach','bestine_iv_cliff_hunter','bestine_iv_cliff_hunter','bestine_iv_cliff_hunter'),
 ('bestine_iv','forgotten_tide_instrument_post','bestine_iv_survey_droid','bestine_iv_survey_droid','bestine_iv_survey_droid'),
 ('bestine_iv','deep_coastal_hollows','bestine_iv_shore_crab','bestine_iv_cliff_hunter','bestine_iv_cliff_hunter'),
 ('bestine_iv','remote_island_raider_camp','bestine_iv_expedition_raider','bestine_iv_survey_droid','bestine_iv_expedition_raider'),
 ('denon','service_cargo_landing','denon_cargo_thief','denon_loader_droid','denon_cargo_thief'),
 ('denon','lower_storage_lanes','denon_cargo_thief','denon_gang_enforcer','denon_cargo_thief'),
 ('denon','idle_distribution_hall','denon_loader_droid','denon_loader_droid','denon_loader_droid'),
 ('denon','closed_utility_passages','denon_utility_droid','denon_utility_droid','denon_utility_droid'),
 ('denon','seized_freight_courtyards','denon_cargo_thief','denon_gang_enforcer','denon_loader_droid'),
 ('denon','underground_transfer_gallery','denon_cargo_thief','denon_utility_droid','denon_gang_enforcer'),
 ('denon','outlaw_equipment_workshops','denon_gang_enforcer','denon_loader_droid','denon_gang_enforcer'),
 ('denon','abandoned_service_stack','denon_utility_droid','denon_loader_droid','denon_utility_droid'),
 ('denon','hidden_warehouse_vaults','denon_cargo_thief','denon_gang_enforcer','denon_gang_enforcer'),
 ('denon','fortified_logistics_block','denon_gang_enforcer','denon_loader_droid','denon_utility_droid'),
 ('mygeeto','sheltered_ice_landing','mygeeto_crystal_raider','mygeeto_tunnel_droid','mygeeto_crystal_raider'),
 ('mygeeto','crystal_ridge_path','mygeeto_ice_stalker','mygeeto_ice_stalker','mygeeto_ice_stalker'),
 ('mygeeto','abandoned_mountain_store','mygeeto_crystal_raider','mygeeto_tunnel_droid','mygeeto_tunnel_droid'),
 ('mygeeto','wind_sheltered_ice_gorge','mygeeto_ice_stalker','mygeeto_ice_stalker','mygeeto_ice_stalker'),
 ('mygeeto','idle_boring_workshops','mygeeto_tunnel_droid','mygeeto_tunnel_droid','mygeeto_tunnel_droid'),
 ('mygeeto','raider_crystal_depot','mygeeto_crystal_raider','mygeeto_tunnel_droid','mygeeto_crystal_raider'),
 ('mygeeto','ice_worm_shelf_approach','mygeeto_ice_stalker','mygeeto_ice_worm','mygeeto_ice_worm'),
 ('mygeeto','deep_crystal_service_gallery','mygeeto_crystal_raider','mygeeto_tunnel_droid','mygeeto_tunnel_droid'),
 ('mygeeto','frozen_worm_hollows','mygeeto_ice_worm','mygeeto_ice_worm','mygeeto_ice_worm'),
 ('mygeeto','remote_ice_ridge_dens','mygeeto_ice_stalker','mygeeto_ice_worm','mygeeto_ice_worm'),
 ('muunilinst','mineral_supply_landing','muunilinst_metal_raider','muunilinst_ore_droid','muunilinst_metal_raider'),
 ('muunilinst','grassy_foothill_track','muunilinst_plain_stalker','muunilinst_mountain_hunter','muunilinst_plain_stalker'),
 ('muunilinst','abandoned_metal_storehouses','muunilinst_metal_raider','muunilinst_ore_droid','muunilinst_metal_raider'),
 ('muunilinst','wooded_valley_trail','muunilinst_plain_stalker','muunilinst_mountain_hunter','muunilinst_plain_stalker'),
 ('muunilinst','idle_ore_workyard','muunilinst_ore_droid','muunilinst_ore_droid','muunilinst_ore_droid'),
 ('muunilinst','raider_transfer_compound','muunilinst_metal_raider','muunilinst_ore_droid','muunilinst_metal_raider'),
 ('muunilinst','jagged_mountain_pass','muunilinst_mountain_hunter','muunilinst_mountain_hunter','muunilinst_mountain_hunter'),
 ('muunilinst','sheltered_mineral_coast','muunilinst_metal_raider','muunilinst_ore_droid','muunilinst_metal_raider'),
 ('muunilinst','deep_woodland_hollows','muunilinst_plain_stalker','muunilinst_mountain_hunter','muunilinst_mountain_hunter'),
 ('muunilinst','fortified_metal_depot','muunilinst_metal_raider','muunilinst_ore_droid','muunilinst_metal_raider')
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
WHERE p.code IN ('malastare','kessel','eriadu','sluis_van','fondor','bestine_iv','denon','mygeeto','muunilinst') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
