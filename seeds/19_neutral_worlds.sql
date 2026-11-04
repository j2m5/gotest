-- Third neutral-world batch. Ownership/access remain independent of content.
-- Early Great Galactic War: authored districts and roles, informed by Legends.
-- No late Ilum invasion, Revanite campaign, Mandalorian Excision or Vong devastation.
-- Cathar's surviving underground kiltik colonies are an explicit game assumption.
-- No moon locations, Force effects, crystal mining or released spirits are added.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('ilum_ice_stalker','Хищник ледяных троп','Прячется среди снежных уступов и скальных ниш Илума. Выслеживает добычу у защищённых от ветра переходов.','beast'),
 ('ilum_cave_crawler','Пещерный ползун Илума','Обитает в тёмных расщелинах под ледяной коркой. Нападает на существ, приближающихся к его укрытиям среди камней.','beast'),
 ('ilum_gorgodon','Горгодон','Крупный зверь бродит по снежным низинам и горным склонам Илума. Находит добычу по запаху и звуку, защищает свои укрытия.','beast'),
 ('ilum_survey_droid','Повреждённый исследовательский дроид','Остался на старом посту наблюдения за ледяными пещерами. Сбойные датчики заставляют его применять рабочие инструменты против посетителей.','droid'),
 ('rishi_grophet','Дикий грофет','Роется в почве тропических низин Риши в поисках пищи. Защищается мощными ударами головы, если чужак подходит слишком близко.','beast'),
 ('rishi_jungle_stalker','Хищник островных джунглей','Держится у густых зарослей и каменных ложбин Риши. Подстерегает добычу на тропах между ручьями и лесными укрытиями.','beast'),
 ('rishi_cargo_raider','Налётчик островных пристаней','Грабит небольшие суда и захватывает припасы у перевозчиков. Охраняет нелегальные склады в удалённых бухтах.','humanoid'),
 ('rishi_service_droid','Сбойный причальный техдроид','Обслуживал грузовые площадки и старые топливные узлы. Повреждённая программа заставляет его атаковать движущиеся объекты рабочими захватами.','droid'),
 ('mandalore_strill','Дикий стрилл','Шестиногий хищник охотится в лесах Мандалора. Преследует добычу среди стволов и защищает своё логово.','beast'),
 ('mandalore_rogue_mercenary','Наёмник разбойничьей группы','Вооружённый отступник служит преступникам, нападающим на поселения и караваны. Охраняет захваченные мастерские и склады.','humanoid'),
 ('mandalore_ore_raider','Похититель мандалорской руды','Перехватывает минеральные грузы на удалённых дорогах. Прячет добычу в покинутых дворах и старых выработках.','humanoid'),
 ('mandalore_forge_droid','Повреждённый кузнечный дроид','Остался в закрытой ремесленной мастерской после остановки работ. Принимает посетителей за помехи и применяет тяжёлые инструменты.','droid'),
 ('concord_dawn_plains_stalker','Хищник равнин Конкорд-Доуна','Бродит по травянистым низинам и лесным окраинам. Подстерегает добычу у водопоев и заброшенных хозяйственных дорог.','beast'),
 ('concord_dawn_supply_raider','Грабитель фермерских караванов','Нападает на перевозчиков продовольствия и сельского оборудования. Охраняет награбленное в покинутых фермерских дворах.','humanoid'),
 ('concord_dawn_harvest_droid','Сбойный полевой дроид','Продолжает работать на оставленных без ухода угодьях. Неисправные датчики направляют уборочные инструменты на живые цели.','droid'),
 ('concord_dawn_mercenary','Наёмный охранник разбойников','Защищает тайные базы банды, грабящей сельские поселения. Контролирует склады оружия и подходы к укреплённым стоянкам.','humanoid'),
 ('cathar_grassland_stalker','Хищник катарских саванн','Скрывается в высокой траве у каменистых гряд. Подбирается к добыче возле звериных троп и небольших водоёмов.','beast'),
 ('cathar_kiltik','Килтик','Крупный всеядный жук из уцелевшей подземной колонии. Защищает гнездовые проходы и выходит кормиться на соседние равнины.','beast'),
 ('cathar_supply_raider','Похититель грузов переселенцев','Грабит снабжение возвращающихся на Катар общин. Охраняет украденные припасы в пустующих хозяйственных постройках.','humanoid'),
 ('cathar_restoration_droid','Повреждённый восстановительный дроид','Расчищал старые площадки и ремонтировал подъёмные настилы. После отказа датчиков применяет рабочие инструменты против любого подвижного объекта.','droid'),
 ('myrkr_vornskr','Ворнскр','Лесной хищник Миркра скрывается в густой тени. Выслеживает добычу среди корней и атакует острыми зубами и гибким хвостом.','beast'),
 ('myrkr_root_stalker','Хищник корневых чащ','Занимает каменные полости и укрытия под деревьями Миркра. Подстерегает путников у узких лесных проходов.','beast'),
 ('myrkr_cargo_raider','Налётчик лесных убежищ','Использует глухие районы Миркра для хранения похищенных грузов. Защищает временные стоянки и тайные посадочные площадки.','humanoid'),
 ('myrkr_survey_droid','Сбойный лесной техдроид','Обслуживал оборудование небольших экспедиций среди помех местных лесов. Повреждённые системы распознавания заставляют его атаковать посторонних.','droid'),
 ('iridonia_canyon_stalker','Хищник иридонских каньонов','Держится у сухих скальных укрытий и защищённых долин. Подстерегает добычу на узких тропах над опасными низинами.','beast'),
 ('iridonia_supply_raider','Налётчик каньонных дорог','Грабит караваны между иридонскими поселениями. Охраняет захваченные товары в покинутых грузовых нишах.','humanoid'),
 ('iridonia_tunnel_droid','Повреждённый иридонский проходчик','Расчищал горные выработки и служебные тоннели. После сбоя применяет буровые инструменты против движущихся целей.','droid'),
 ('iridonia_outlaw_guard','Охранник скальных тайников','Нанят преступной группой для защиты складов в труднодоступных каньонах. Перекрывает проходы и атакует приближающихся свидетелей.','humanoid'),
 ('ambria_hssiss','Хссисс','Опасный зверь обитает у озера Натт и в соседних пустошах Амбрии. Скрывается среди камней и атакует существ, приблизившихся к береговым укрытиям.','beast'),
 ('ambria_wasteland_stalker','Хищник амбрийских пустошей','Держится в каменистых ложбинах вдали от редких стоянок. Подстерегает добычу у защищённых проходов и старых выработок.','beast'),
 ('ambria_relic_raider','Мародёр амбрийских стоянок','Грабит малые экспедиции, прибывающие исследовать Амбрию. Прячет находки и украденное оборудование в заброшенных горных постройках.','humanoid'),
 ('ambria_survey_droid','Сбойный разведочный дроид','Остался у старых геологических постов и шахтных входов. Повреждённые инструкции заставляют его защищать пустые площадки от любых посетителей.','droid')
ON CONFLICT (code) DO NOTHING;

-- Ilum and Rishi: remote ice caves and tropical coastal routes.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('ilum','sheltered_ice_landing','Укрытая ледяная площадка','Скальный выступ защищает небольшой посадочный настил от снежных наносов. Рядом сохранилось брошенное исследовательское оборудование, к которому подбираются местные звери.',1,1),
 ('ilum','glacier_edge_trail','Тропа вдоль ледника','Узкий проход огибает высокую ледяную стену. В каменных нишах прячутся хищники, а под снежной коркой начинаются тёмные расщелины.',2,6),
 ('ilum','old_observation_post','Старый наблюдательный пост','Небольшой служебный комплекс стоит у спуска к пещерам. Повреждённые дроиды продолжают обходить пустые помещения и площадки приборов.',3,11),
 ('ilum','frost_cave_approach','Подступы к морозным пещерам','Скальные ворота ведут в защищённые от ветра проходы. Среди ледяных наростов скрываются ползуны и звери, ищущие добычу у входов.',4,16),
 ('ilum','snowbound_hollow','Заснеженная котловина','Высокие уступы окружают долину с редкой открытой растительностью. Местные хищники держатся возле природных укрытий и защищённых троп.',5,21),
 ('ilum','abandoned_instrument_galleries','Заброшенные измерительные галереи','В сухих боковых пещерах остались опоры датчиков и служебные настилы. Неисправные машины работают рядом с занятыми зверями разломами.',6,26),
 ('ilum','gorgodon_lowlands','Низины горгодонов','Снежные холмы переходят в широкие каменистые ложбины. Здесь горгодоны устраивают укрытия и выходят на охоту вдоль подножий скал.',7,31),
 ('ilum','crystal_arch_passages','Проходы кристальных арок','В стенах глубоких пещер мерцают кристальные жилы. Узкие дорожки проходят между природными арками и гнездовыми расщелинами ползунов.',8,36),
 ('ilum','deep_frozen_chambers','Глубокие промёрзшие залы','Каменные полости соединены ледяными переходами и скрытыми лазами. В них укрываются опасные звери, редко покидающие тёмные глубины.',9,41),
 ('ilum','echoing_ice_gorge','Ущелье ледяного эха','Высокие стены отражают звуки движения по снежному дну. Крупные горгодоны занимают боковые укрытия и охраняют подходы к своим логовищам.',10,46),
 ('rishi','coastal_cargo_landing','Береговая грузовая площадка','Деревянные настилы выходят к небольшому причалу у края джунглей. Налётчики используют пустующие навесы для хранения похищенных припасов.',1,1),
 ('rishi','coastal_root_paths','Прибрежные корневые тропы','Густые заросли отделяют песчаный берег от влажных низин. Грофеты ищут корм среди корней, а хищники прячутся в лесной тени.',2,6),
 ('rishi','old_fuel_pier','Старый топливный причал','Ржавые трубы соединяют резервуары с заброшенным погрузочным настилом. Преступники присвоили уцелевшее оборудование и оставили охрану сбойным машинам.',3,11),
 ('rishi','jungle_stream_ford','Брод лесного ручья','Каменная отмель позволяет пересечь поток без погружения в глубокие заводи. У воды держатся местные звери, защищающие привычные тропы.',4,16),
 ('rishi','grophet_grove','Роща диких грофетов','Под широкими кронами много рыхлой почвы и густого подлеска. Грофеты занимают кормовые участки и яростно отгоняют приблизившихся чужаков.',5,21),
 ('rishi','raider_supply_yards','Дворы снабжения налётчиков','За частоколом скрыты краденые контейнеры и ремонтные навесы. Вооружённая группа готовит здесь снаряжение для нападений на перевозчиков.',6,26),
 ('rishi','wooded_foothill_pass','Перевал лесистых предгорий','Тропа поднимается от низин к каменным уступам под густой зеленью. Местные хищники используют боковые ниши, а грофеты занимают защищённые площадки.',7,31),
 ('rishi','idle_refinery_workshops','Мастерские старого переработчика','Закрытый топливный комплекс постепенно зарастает тропической растительностью. Налётчики используют его склады, пока повреждённые дроиды обслуживают пустые цеха.',8,36),
 ('rishi','hidden_lagoon_walkways','Настилы тайной лагуны','Прочные мостки соединяют небольшие склады вдоль укрытого берега. Преступники перегружают добычу под защитой скал и густых зарослей.',9,41),
 ('rishi','fortified_cove_hideout','Укреплённое убежище бухты','Покинутый береговой комплекс превращён в главную базу местной банды. Вооружённые дозоры и переделанная техника перекрывают дорогу к грузовым складам.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Mandalore and Concord Dawn before the much later Excision.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('mandalore','forest_trade_landing','Торговая площадка у леса','Грузовой настил расположен на окраине старой ремесленной дороги. В пустующих навесах скрываются похитители руды и нанятые ими разбойники.',1,1),
 ('mandalore','veshok_grove_trail','Тропа рощи вешок','Высокие деревья закрывают каменистые ложбины густой тенью. Дикие стриллы держатся у звериных переходов и защищают свои укрытия.',2,6),
 ('mandalore','old_ore_yards','Старые рудные дворы','Закрытые площадки хранения примыкают к покинутым выработкам. Преступники сортируют похищенные грузы рядом с неисправной рабочей техникой.',3,11),
 ('mandalore','hill_caravan_route','Караванный путь холмов','Дорога огибает травянистые склоны и лесные овраги. Разбойничьи группы наблюдают за перевозчиками из старых придорожных укрытий.',4,16),
 ('mandalore','shuttered_forge_annex','Закрытый кузнечный корпус','Пустующие мастерские соединены дворами для металла и инструментов. Сбойные кузнечные дроиды продолжают опасные рабочие циклы среди тайников грабителей.',5,21),
 ('mandalore','woodland_stream_paths','Лесные тропы у ручьёв','Каменные отмели и корни образуют проходы между влажными низинами. Стриллы занимают защищённые участки и преследуют добычу вдоль воды.',6,26),
 ('mandalore','hidden_mercenary_camp','Скрытый лагерь разбойников','За лесным гребнем устроены ремонтные навесы и оружейные тайники. Наёмники преступной группы охраняют запасы и готовят новые налёты.',7,31),
 ('mandalore','deep_ore_galleries','Глубокие рудные галереи','Заброшенные штольни уходят под старый каменный карьер. Похитители руды используют боковые залы вместе с присвоенными рабочими машинами.',8,36),
 ('mandalore','abandoned_fort_workshops','Мастерские покинутого укрепления','Уцелевшие каменные стены окружают старые ремонтные дворы. Разбойники восстановили часть оборудования и перекрыли подходы вооружёнными постами.',9,41),
 ('mandalore','raider_arms_store','Оружейная база налётчиков','Укреплённый склад скрыт за цепью нежилых хозяйственных корпусов. Преступная сеть собрала здесь основные запасы, наёмную охрану и переделанные машины.',10,46),
 ('concord_dawn','farm_supply_landing','Площадка фермерского снабжения','Небольшой грузовой настил стоит у края полей и старых сараев. Грабители прячут добычу среди брошенных хозяйственных агрегатов.',1,1),
 ('concord_dawn','dry_field_tracks','Дороги сухих полей','Пыльные колеи соединяют отдалённые угодья и небольшие рощи. Налётчики наблюдают за перевозчиками, а полевые дроиды бродят по оставленным участкам.',2,6),
 ('concord_dawn','woodland_waterholes','Водопои лесной окраины','Неглубокие водоёмы скрыты между деревьями и травянистыми склонами. Местные хищники занимают подходы к воде вдали от жилых дворов.',3,11),
 ('concord_dawn','abandoned_grain_yards','Покинутые зерновые дворы','Пустые хранилища окружены разгрузочными площадками и навесами. Банда присвоила остатки припасов и использует сбойную технику для охраны.',4,16),
 ('concord_dawn','rocky_ranch_road','Каменистая дорога пастбищ','Тропа проходит по сухим грядам между лесными участками и открытыми равнинами. Хищники и грабители используют боковые ложбины для засад.',5,21),
 ('concord_dawn','seized_equipment_sheds','Захваченные склады инвентаря','Старые сельские мастерские заняты похитителями оборудования. Внутри спрятаны украденные детали, оружие и несколько неисправных машин.',6,26),
 ('concord_dawn','wild_grassland_hollows','Лощины диких равнин','Высокая трава скрывает естественные укрытия на склонах. Крупные местные хищники защищают логовища и выходы к водопоям.',7,31),
 ('concord_dawn','outlaw_homestead','Разбойничья база в старой усадьбе','Покинутое хозяйство превращено в укреплённую стоянку. Наёмная охрана контролирует склады награбленного и ремонтные дворы.',8,36),
 ('concord_dawn','hidden_caravan_store','Тайный склад караванной добычи','Изолированный комплекс спрятан за лесной полосой и каменными оградами. Здесь банда хранит наиболее ценные грузы и готовит их к незаконной продаже.',9,41),
 ('concord_dawn','fortified_supply_compound','Укреплённый узел разбойничьего снабжения','Несколько старых хозяйственных корпусов объединены защитными ограждениями. Вооружённые наёмники и переделанная техника охраняют главные запасы преступной группы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Cathar's returning communities and Myrkr's dense forests.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('cathar','city_tree_supply_landing','Площадка у городского дерева','Грузовой настил расположен рядом с восстанавливаемыми подъёмными путями. В закрытых хозяйственных помещениях скрываются похитители снабжения и повреждённые машины.',1,1),
 ('cathar','tallgrass_paths','Тропы высокой травы','Саванна начинается сразу за редкими хозяйственными оградами. Местные хищники прячутся среди травы и каменных выступов.',2,6),
 ('cathar','old_resettlement_stores','Старые склады переселенцев','Часть временных построек уже покинута после переноса лагеря. Грабители заняли пустые кладовые и охраняют похищенные припасы.',3,11),
 ('cathar','rocky_savanna_ridges','Скалистые гряды саванны','Узкие проходы поднимаются над травянистыми низинами. Хищники устраивают укрытия в расщелинах, а налётчики наблюдают за дальними дорогами.',4,16),
 ('cathar','damaged_lift_yards','Дворы повреждённых подъёмников','Старые опоры окружают закрытые ремонтные площадки. Восстановительные дроиды работают со сбоями рядом с тайниками похитителей оборудования.',5,21),
 ('cathar','raided_convoy_station','Захваченная караванная стоянка','Покинутый перевалочный пункт стал базой грабителей снабжения. Между навесами собраны чужие контейнеры и присвоенная рабочая техника.',6,26),
 ('cathar','kiltik_tunnel_approaches','Подступы к тоннелям килтиков','Среди истощённой почвы открываются входы в уцелевшую подземную колонию. Крупные жуки охраняют проходы и выходят за пищей на соседние равнины.',7,31),
 ('cathar','abandoned_recovery_workshops','Покинутые восстановительные мастерские','Закрытые цеха стоят у старых транспортных троп. Преступники хранят здесь украденные инструменты и используют повреждённых дроидов для охраны.',8,36),
 ('cathar','stripped_hive_plains','Истощённые равнины у колонии','Растительность вокруг гнездовых выходов почти исчезла. Килтики пересекают открытые участки, а саванные хищники держатся у дальних каменных гряд.',9,41),
 ('cathar','deep_kiltik_passages','Глубокие проходы колонии','Просторные подземные тоннели соединяют несколько гнездовых залов. Крупные килтики защищают внутренние пути и не подпускают чужаков к глубинам колонии.',10,46),
 ('myrkr','forest_cargo_clearing','Грузовая поляна в лесу','Небольшой просвет между деревьями пригоден для посадки малого транспорта. Налётчики спрятали припасы в старых навесах, оставшихся от временной стоянки.',1,1),
 ('myrkr','narrow_woodland_tracks','Узкие лесные тропы','Густые ветви скрывают проходы между скалами и корнями. Местные звери охраняют укрытия и подстерегают добычу в тени.',2,6),
 ('myrkr','abandoned_sensor_post','Покинутый пост датчиков','Среди деревьев стоят накренившиеся опоры исследовательского оборудования. Сбойные техдроиды продолжают обходить площадку, занятую похитителями грузов.',3,11),
 ('myrkr','fallen_log_crossing','Переход поваленных стволов','Массивные деревья образовали сухой путь над лесной низиной. Ворнскры и другие хищники держатся у узких выходов между корнями.',4,16),
 ('myrkr','hidden_supply_camp','Скрытая стоянка снабжения','В глубине зарослей устроены навесы и небольшие склады. Вооружённая группа использует их для сокрытия добычи от проходящих экспедиций.',5,21),
 ('myrkr','olbio_root_paths','Тропы корней олбио','Толстые корни переплетаются под сомкнутыми кронами. Ворнскры проходят здесь на охоту, а небольшие хищники занимают каменные полости.',6,26),
 ('myrkr','shaded_river_trail','Тропа тенистой реки','Лесная дорога следует по высокому берегу вдоль густой растительности. Хищники занимают ложбины и укрытия у воды.',7,31),
 ('myrkr','forgotten_survey_huts','Забытые исследовательские хижины','Пустующие модули почти скрыты зеленью и старыми поваленными деревьями. Налётчики используют их склады рядом с повреждёнными служебными машинами.',8,36),
 ('myrkr','vornskr_forest_dens','Логовища ворнскров','В труднопроходимой чаще сохранились широкие полости под корнями. Ворнскры защищают свои укрытия и выходы к охотничьим тропам.',9,41),
 ('myrkr','concealed_landing_compound','Тайная база у посадочной поляны','Укреплённые временные постройки скрыты глубоко в лесу. Налётчики собрали здесь ценные грузы, вооружённую охрану и переделанное оборудование.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Iridonia's dry canyon routes and Ambria's wastelands outside Lake Natth.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('iridonia','canyon_supply_landing','Грузовая терраса каньона','Высокие скальные стены защищают небольшой посадочный настил. Налётчики заняли пустующие кладовые у служебной дороги.',1,1),
 ('iridonia','dry_ledge_paths','Тропы сухих уступов','Каменные полки проходят над глубокими низинами. В боковых нишах скрываются местные хищники и наблюдатели преступной группы.',2,6),
 ('iridonia','old_valley_storehouses','Старые склады долины','В защищённом участке каньона сохранились закрытые хозяйственные корпуса. Грабители хранят здесь товары, отнятые у проходящих караванов.',3,11),
 ('iridonia','acid_basin_overlook','Уступ над кислотной низиной','Сухой путь идёт по прочному каменному краю высоко над опасными водоёмами. Хищники занимают скальные полости вдоль этой единственной удобной тропы.',4,16),
 ('iridonia','wind_sheltered_gorge','Укрытое от ветра ущелье','Повороты высоких стен защищают внутренний проход от сильных порывов. Налётчики устроили здесь стоянки и наблюдательные позиции.',5,21),
 ('iridonia','abandoned_boring_yards','Покинутые дворы проходчиков','Старые выработки начинаются за закрытыми ремонтными площадками. Сбойные машины продолжают движение среди тайников похитителей оборудования.',6,26),
 ('iridonia','rock_predator_hollows','Скальные логовища хищников','Цепь сухих расщелин уходит в защищённые полости каньона. Крупные звери охраняют входы и выходы к охотничьим тропам.',7,31),
 ('iridonia','deep_service_tunnels','Глубокие служебные тоннели','Закрытые грузовые проходы соединяют покинутые участки добычи. Преступники используют их для перевозки краденого под защитой повреждённых машин.',8,36),
 ('iridonia','stolen_machine_vaults','Хранилища похищенных машин','Укреплённые помещения скрыты внутри старого скального комплекса. Наёмная охрана защищает ценные детали и подготовленное к вывозу оборудование.',9,41),
 ('iridonia','outlaw_cliff_fort','Скальное укрепление налётчиков','Преступная сеть заняла изолированный комплекс над узкой дорогой. Охранники и тяжёлые рабочие автоматы перекрывают все подходы к основным складам.',10,46),
 ('ambria','wasteland_survey_landing','Посадочная площадка пустошей','Каменная терраса пригодна для остановки небольшой экспедиции. Среди брошенных приборов скрываются мародёры и повреждённые разведочные машины.',1,1),
 ('ambria','mineral_ridge_trail','Тропа минеральной гряды','Сухой проход огибает обнажения породы и старые геологические отметки. Местные хищники используют скальные ниши для укрытия.',2,6),
 ('ambria','old_geological_post','Старый геологический пост','Несколько пустующих модулей сохранились у заброшенных разведочных выработок. Мародёры прячут здесь оборудование рядом со сбойными служебными дроидами.',3,11),
 ('ambria','wind_worn_hollows','Котловины выветренных скал','Каменные стенки образуют цепь защищённых от ветра низин. Звери занимают боковые проходы, а грабители следят за редкими путниками.',4,16),
 ('ambria','abandoned_mining_yards','Покинутые горные дворы','От старых попыток добычи остались склады инструмента и грузовые опоры. Мародёры используют пустые помещения и присвоенную разведочную технику.',5,21),
 ('ambria','sparse_green_valley','Долина редкой зелени','Среди каменных ложбин появились участки травы и низких растений. Местные звери держатся у этих кормовых мест и защищают соседние укрытия.',6,26),
 ('ambria','natth_stone_approach','Каменные подступы к Натту','Сухая тропа следует вдоль скальных гряд к внешним береговым участкам. Хссисы прячутся среди камней и не подпускают чужаков к своим укрытиям.',7,31),
 ('ambria','sealed_rock_store','Тайник в старом скальном складе','Закрытые служебные помещения сохранились у заброшенной штольни. Мародёры собрали здесь ценные находки и запустили повреждённых дроидов для охраны.',8,36),
 ('ambria','natth_outer_banks','Внешние берега Натта','Каменистые полосы суши тянутся вдоль тёмной воды. Хссисы занимают береговые расщелины и охраняют проходы между ними.',9,41),
 ('ambria','hssiss_shore_hollows','Береговые логовища хссиссов','Высокие скальные стенки окружают несколько сухих укрытий у озера. Здесь держатся опасные звери, защищающие свои логовища и выходы в пустоши.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('ilum','sheltered_ice_landing','ilum_ice_stalker','ilum_survey_droid','ilum_ice_stalker'),
 ('ilum','glacier_edge_trail','ilum_ice_stalker','ilum_cave_crawler','ilum_ice_stalker'),
 ('ilum','old_observation_post','ilum_survey_droid','ilum_survey_droid','ilum_survey_droid'),
 ('ilum','frost_cave_approach','ilum_cave_crawler','ilum_ice_stalker','ilum_cave_crawler'),
 ('ilum','snowbound_hollow','ilum_ice_stalker','ilum_cave_crawler','ilum_ice_stalker'),
 ('ilum','abandoned_instrument_galleries','ilum_survey_droid','ilum_cave_crawler','ilum_survey_droid'),
 ('ilum','gorgodon_lowlands','ilum_ice_stalker','ilum_gorgodon','ilum_gorgodon'),
 ('ilum','crystal_arch_passages','ilum_cave_crawler','ilum_cave_crawler','ilum_cave_crawler'),
 ('ilum','deep_frozen_chambers','ilum_cave_crawler','ilum_gorgodon','ilum_gorgodon'),
 ('ilum','echoing_ice_gorge','ilum_ice_stalker','ilum_gorgodon','ilum_gorgodon'),
 ('rishi','coastal_cargo_landing','rishi_cargo_raider','rogue_loader','rishi_cargo_raider'),
 ('rishi','coastal_root_paths','rishi_grophet','rishi_jungle_stalker','rishi_grophet'),
 ('rishi','old_fuel_pier','rishi_service_droid','rishi_cargo_raider','rishi_service_droid'),
 ('rishi','jungle_stream_ford','rishi_grophet','rishi_jungle_stalker','rishi_jungle_stalker'),
 ('rishi','grophet_grove','rishi_grophet','rishi_grophet','rishi_grophet'),
 ('rishi','raider_supply_yards','rishi_cargo_raider','rishi_service_droid','rishi_cargo_raider'),
 ('rishi','wooded_foothill_pass','rishi_grophet','rishi_jungle_stalker','rishi_jungle_stalker'),
 ('rishi','idle_refinery_workshops','rishi_service_droid','rishi_cargo_raider','rishi_service_droid'),
 ('rishi','hidden_lagoon_walkways','rishi_cargo_raider','rishi_service_droid','rishi_cargo_raider'),
 ('rishi','fortified_cove_hideout','rishi_cargo_raider','rishi_service_droid','rishi_cargo_raider'),
 ('mandalore','forest_trade_landing','mandalore_ore_raider','mandalore_rogue_mercenary','mandalore_ore_raider'),
 ('mandalore','veshok_grove_trail','mandalore_strill','mandalore_strill','mandalore_strill'),
 ('mandalore','old_ore_yards','mandalore_ore_raider','mandalore_forge_droid','mandalore_ore_raider'),
 ('mandalore','hill_caravan_route','mandalore_ore_raider','mandalore_rogue_mercenary','mandalore_rogue_mercenary'),
 ('mandalore','shuttered_forge_annex','mandalore_forge_droid','mandalore_ore_raider','mandalore_forge_droid'),
 ('mandalore','woodland_stream_paths','mandalore_strill','mandalore_strill','mandalore_strill'),
 ('mandalore','hidden_mercenary_camp','mandalore_rogue_mercenary','mandalore_ore_raider','mandalore_rogue_mercenary'),
 ('mandalore','deep_ore_galleries','mandalore_ore_raider','mandalore_forge_droid','mandalore_forge_droid'),
 ('mandalore','abandoned_fort_workshops','mandalore_rogue_mercenary','mandalore_forge_droid','mandalore_rogue_mercenary'),
 ('mandalore','raider_arms_store','mandalore_ore_raider','mandalore_rogue_mercenary','mandalore_forge_droid'),
 ('concord_dawn','farm_supply_landing','concord_dawn_supply_raider','concord_dawn_harvest_droid','concord_dawn_supply_raider'),
 ('concord_dawn','dry_field_tracks','concord_dawn_supply_raider','concord_dawn_harvest_droid','concord_dawn_supply_raider'),
 ('concord_dawn','woodland_waterholes','concord_dawn_plains_stalker','concord_dawn_plains_stalker','concord_dawn_plains_stalker'),
 ('concord_dawn','abandoned_grain_yards','concord_dawn_supply_raider','concord_dawn_harvest_droid','concord_dawn_mercenary'),
 ('concord_dawn','rocky_ranch_road','concord_dawn_plains_stalker','concord_dawn_supply_raider','concord_dawn_mercenary'),
 ('concord_dawn','seized_equipment_sheds','concord_dawn_supply_raider','concord_dawn_harvest_droid','concord_dawn_harvest_droid'),
 ('concord_dawn','wild_grassland_hollows','concord_dawn_plains_stalker','concord_dawn_plains_stalker','concord_dawn_plains_stalker'),
 ('concord_dawn','outlaw_homestead','concord_dawn_supply_raider','concord_dawn_mercenary','concord_dawn_mercenary'),
 ('concord_dawn','hidden_caravan_store','concord_dawn_supply_raider','concord_dawn_mercenary','concord_dawn_mercenary'),
 ('concord_dawn','fortified_supply_compound','concord_dawn_supply_raider','concord_dawn_harvest_droid','concord_dawn_mercenary'),
 ('cathar','city_tree_supply_landing','cathar_supply_raider','cathar_restoration_droid','cathar_supply_raider'),
 ('cathar','tallgrass_paths','cathar_grassland_stalker','cathar_grassland_stalker','cathar_grassland_stalker'),
 ('cathar','old_resettlement_stores','cathar_supply_raider','cathar_restoration_droid','cathar_supply_raider'),
 ('cathar','rocky_savanna_ridges','cathar_grassland_stalker','cathar_supply_raider','cathar_grassland_stalker'),
 ('cathar','damaged_lift_yards','cathar_restoration_droid','cathar_supply_raider','cathar_restoration_droid'),
 ('cathar','raided_convoy_station','cathar_supply_raider','cathar_restoration_droid','cathar_supply_raider'),
 ('cathar','kiltik_tunnel_approaches','cathar_kiltik','cathar_kiltik','cathar_kiltik'),
 ('cathar','abandoned_recovery_workshops','cathar_supply_raider','cathar_restoration_droid','cathar_supply_raider'),
 ('cathar','stripped_hive_plains','cathar_kiltik','cathar_grassland_stalker','cathar_kiltik'),
 ('cathar','deep_kiltik_passages','cathar_kiltik','cathar_kiltik','cathar_kiltik'),
 ('myrkr','forest_cargo_clearing','myrkr_cargo_raider','myrkr_survey_droid','myrkr_cargo_raider'),
 ('myrkr','narrow_woodland_tracks','myrkr_root_stalker','myrkr_vornskr','myrkr_root_stalker'),
 ('myrkr','abandoned_sensor_post','myrkr_survey_droid','myrkr_cargo_raider','myrkr_survey_droid'),
 ('myrkr','fallen_log_crossing','myrkr_root_stalker','myrkr_vornskr','myrkr_vornskr'),
 ('myrkr','hidden_supply_camp','myrkr_cargo_raider','myrkr_survey_droid','myrkr_cargo_raider'),
 ('myrkr','olbio_root_paths','myrkr_root_stalker','myrkr_vornskr','myrkr_vornskr'),
 ('myrkr','shaded_river_trail','myrkr_root_stalker','myrkr_vornskr','myrkr_vornskr'),
 ('myrkr','forgotten_survey_huts','myrkr_cargo_raider','myrkr_survey_droid','myrkr_survey_droid'),
 ('myrkr','vornskr_forest_dens','myrkr_vornskr','myrkr_vornskr','myrkr_vornskr'),
 ('myrkr','concealed_landing_compound','myrkr_cargo_raider','myrkr_survey_droid','myrkr_cargo_raider'),
 ('iridonia','canyon_supply_landing','iridonia_supply_raider','rogue_loader','iridonia_supply_raider'),
 ('iridonia','dry_ledge_paths','iridonia_canyon_stalker','iridonia_supply_raider','iridonia_canyon_stalker'),
 ('iridonia','old_valley_storehouses','iridonia_supply_raider','iridonia_tunnel_droid','iridonia_supply_raider'),
 ('iridonia','acid_basin_overlook','iridonia_canyon_stalker','iridonia_canyon_stalker','iridonia_canyon_stalker'),
 ('iridonia','wind_sheltered_gorge','iridonia_supply_raider','iridonia_outlaw_guard','iridonia_supply_raider'),
 ('iridonia','abandoned_boring_yards','iridonia_tunnel_droid','iridonia_supply_raider','iridonia_tunnel_droid'),
 ('iridonia','rock_predator_hollows','iridonia_canyon_stalker','iridonia_canyon_stalker','iridonia_canyon_stalker'),
 ('iridonia','deep_service_tunnels','iridonia_supply_raider','iridonia_tunnel_droid','iridonia_tunnel_droid'),
 ('iridonia','stolen_machine_vaults','iridonia_outlaw_guard','iridonia_tunnel_droid','iridonia_outlaw_guard'),
 ('iridonia','outlaw_cliff_fort','iridonia_supply_raider','iridonia_outlaw_guard','iridonia_tunnel_droid'),
 ('ambria','wasteland_survey_landing','ambria_relic_raider','ambria_survey_droid','ambria_relic_raider'),
 ('ambria','mineral_ridge_trail','ambria_wasteland_stalker','ambria_wasteland_stalker','ambria_wasteland_stalker'),
 ('ambria','old_geological_post','ambria_relic_raider','ambria_survey_droid','ambria_survey_droid'),
 ('ambria','wind_worn_hollows','ambria_wasteland_stalker','ambria_relic_raider','ambria_wasteland_stalker'),
 ('ambria','abandoned_mining_yards','ambria_relic_raider','ambria_survey_droid','ambria_relic_raider'),
 ('ambria','sparse_green_valley','ambria_wasteland_stalker','ambria_wasteland_stalker','ambria_wasteland_stalker'),
 ('ambria','natth_stone_approach','ambria_wasteland_stalker','ambria_hssiss','ambria_hssiss'),
 ('ambria','sealed_rock_store','ambria_relic_raider','ambria_survey_droid','ambria_survey_droid'),
 ('ambria','natth_outer_banks','ambria_hssiss','ambria_hssiss','ambria_hssiss'),
 ('ambria','hssiss_shore_hollows','ambria_hssiss','ambria_hssiss','ambria_hssiss')
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
WHERE p.code IN ('ilum','rishi','mandalore','concord_dawn','cathar','myrkr','iridonia','ambria') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
