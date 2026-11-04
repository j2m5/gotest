-- Fourth neutral-world batch. Ownership/access remain independent of content.
-- Early Great Galactic War: authored districts and roles, informed by Legends.
-- Kwenn's surface yards and Ubrikkia's early industry are explicit game adaptations.
-- No orbital station, High Republic renewal, Clone Wars facilities or late corporations.
-- Klatooine's sacred Fountain is outside combat areas; no extinct Kintan fauna is used.
-- Ground/boardwalk routes do not add swimming, atmospheric hazards or combat effects.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('sleheyron_dock_raider','Налётчик грузовых дворов','Перехватывает контейнеры на окраинах промышленных районов Слейхейрона. Вместе с сообщниками охраняет похищенные грузы.','humanoid'),
 ('sleheyron_furnace_droid','Сбойный термический дроид','Обслуживал оборудование перерабатывающего комплекса. Повреждённые датчики заставляют машину применять рабочие инструменты против посетителей.','droid'),
 ('sleheyron_gang_enforcer','Боевик заводской банды','Вооружённый преступник удерживает захваченные цеха и перевалочные дворы. Перекрывает проходы к складам банды.','humanoid'),
 ('sleheyron_cargo_saboteur','Диверсант грузовых линий','Повреждает оборудование перевозчиков, чтобы присвоить их товары. Прячется среди трубопроводов и служебных галерей.','humanoid'),
 ('toydaria_grabworm','Червь-хвататель','Прячется в илистой воде болот Тойдарии. Бросается на добычу у кромки водорослевых зарослей и низких настилов.','beast'),
 ('toydaria_marsh_predator','Хищник топких заводей','Обитает у болотных озёр и заросших проток. Выслеживает добычу возле сухих опор и небольших островков.','beast'),
 ('toydaria_harvest_raider','Грабитель водорослевых хозяйств','Отнимает припасы у местных сборщиков и перевозчиков. Устраивает тайники на покинутых хозяйственных настилах.','humanoid'),
 ('toydaria_pump_droid','Повреждённый насосный дроид','Остался у заброшенных водозаборных устройств. Из-за сбоя принимает посетителей за препятствия и атакует рабочими захватами.','droid'),
 ('kwenn_cargo_thief','Похититель перевалочных грузов','Вскрывает чужие контейнеры на наземных площадках Квенна. Прячет добычу в закрытых складских секторах.','humanoid'),
 ('kwenn_outlaw_sentry','Часовой контрабандного двора','Служит преступной группе, захватившей часть грузовых построек. Охраняет проходы между её складами и ремонтными навесами.','humanoid'),
 ('kwenn_service_droid','Сбойный дроид грузового узла','Продолжает выполнять старые команды на покинутой площадке. Повреждённые средства распознавания превращают обслуживание в нападение.','droid'),
 ('kwenn_raider_technician','Техник налётчиков','Переоборудует украденные машины и отключает охранные устройства перевозчиков. Защищает подпольные мастерские банды.','humanoid'),
 ('nar_kreeta_tunnel_stalker','Хищник скальных тоннелей','Скрывается в тёмных нишах горных проходов Нар-Криты. Защищает логовища и подстерегает добычу возле выходов на поверхность.','beast'),
 ('nar_kreeta_cargo_raider','Налётчик горных дорог','Перехватывает грузы на узких скальных маршрутах. Держит тайники в покинутых складах и боковых проходах.','humanoid'),
 ('nar_kreeta_mining_droid','Сбойный проходческий дроид','Остался в закрытых выработках и служебных тоннелях. Применяет рабочие инструменты против любого, кто оказывается на пути.','droid'),
 ('nar_kreeta_outlaw_enforcer','Боевик подземного убежища','Охраняет захваченные бандой галереи и склады. Перекрывает тесные входы к внутренним помещениям.','humanoid'),
 ('klatooine_dune_stalker','Хищник песчаных ложбин','Роет укрытия среди сухих склонов и каменных гряд Клатуина. Подстерегает добычу у редких тенистых переходов.','beast'),
 ('klatooine_caravan_raider','Грабитель пустынных караванов','Нападает на перевозчиков припасов вдали от поселений. Устраивает засады у старых стоянок и дорожных складов.','humanoid'),
 ('klatooine_pump_droid','Сбойный дроид водозабора','Обслуживал оборудование заброшенного хозяйственного пункта. Повреждённая программа заставляет его защищать пустые установки от посетителей.','droid'),
 ('klatooine_raider_guard','Охранник караванной банды','Вооружённый преступник удерживает склады захваченных товаров. Защищает тайники и опорные пункты грабителей.','humanoid'),
 ('vodran_dianoga','Болотная дианога','Скрывается в мутной воде Водрана, выставляя над поверхностью глазной стебель. Защищает укрытие и хватает добычу, подошедшую к берегу.','beast'),
 ('vodran_jungle_stalker','Хищник влажных джунглей','Выслеживает добычу среди густых корней и зарослей. Держится возле звериных троп, ведущих к болотным протокам.','beast'),
 ('vodran_river_raider','Грабитель речных пристаней','Перехватывает припасы на удалённых водных маршрутах. Укрывает похищенные грузы на заброшенных настилах.','humanoid'),
 ('vodran_salvage_droid','Сбойный дроид болотных причалов','Разбирал старое оборудование у воды. Влага повредила системы распознавания, и машина атакует посетителей рабочими захватами.','droid'),
 ('kintan_wasteland_predator','Хищник каменных пустошей','Обитает среди сухих гряд и остывших вулканических склонов Кинтана. Выслеживает добычу возле природных укрытий.','beast'),
 ('kintan_forest_stalker','Лесной хищник Кинтана','Скрывается в зарослях у влажных низин и лесистых берегов. Защищает свои логовища среди корней и скальных ниш.','beast'),
 ('kintan_ore_raider','Похититель руды','Грабит перевозчиков на удалённых дорогах и возле выработок. Держит тайники в заброшенных грузовых дворах.','humanoid'),
 ('kintan_quarry_droid','Повреждённый карьерный дроид','Остался на остановленном участке добычи. Сбойная программа заставляет машину применять инструменты против любых движущихся объектов.','droid'),
 ('ubrikkia_assembly_droid','Сбойный сборочный дроид','Продолжает выполнять команды на закрытой производственной линии Убриккии. Повреждённые датчики заставляют его хватать движущиеся объекты.','droid'),
 ('ubrikkia_cargo_raider','Налётчик промышленных складов','Присваивает партии деталей и нападает на перевозчиков. Прячет добычу в покинутых производственных дворах.','humanoid'),
 ('ubrikkia_illegal_guard','Боевик подпольного цеха','Охраняет захваченные мастерские преступной группы. Защищает склады краденого оборудования и тайные погрузочные участки.','humanoid'),
 ('ubrikkia_salvage_droid','Повреждённый разборочный дроид','Разбирал старые механизмы на запасные части. Сбой управляющей программы превратил его инструменты в угрозу для посетителей.','droid')
ON CONFLICT (code) DO NOTHING;

-- Sleheyron.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('sleheyron','outer_freight_landing','Внешний грузовой настил','Посадочная площадка примыкает к цепочке складов. Налётчики скрывают краденые контейнеры между пустующими погрузочными постами.',1,1),
 ('sleheyron','basalt_service_road','Служебная дорога среди базальта','Укреплённый путь проходит между застывшими вулканическими выступами. Банда использует тесные повороты для нападений на грузовые колонны.',2,6),
 ('sleheyron','idle_processing_yard','Двор остановленного переработчика','В ограждённом дворе остались старые резервуары и ремонтные навесы. Диверсанты разбирают оборудование, избегая сбойных машин.',3,11),
 ('sleheyron','pipe_access_galleries','Галереи обслуживания труб','Закрытые переходы тянутся вдоль остывших трубопроводов. Здесь укрываются похитители деталей, а рабочие дроиды продолжают обходы.',4,16),
 ('sleheyron','stolen_fuel_depot','Склад похищенного топлива','Банда устроила перевалочный пункт в заброшенном резервуарном парке. Вооружённые посты охраняют подходы к грузовым площадкам.',5,21),
 ('sleheyron','sealed_furnace_halls','Закрытые термические цеха','Изолированные рабочие коридоры проходят между отключёнными установками. Сбойные дроиды продолжают обслуживание давно остановленных линий.',6,26),
 ('sleheyron','slag_ridge_passage','Проход вдоль шлакового гребня','Техническая дорога огибает высокие насыпи и входит под защитный навес. Боевики прикрывают маршрут перевозки краденого оборудования.',7,31),
 ('sleheyron','abandoned_dispatch_block','Покинутый диспетчерский блок','Защитные стены отделяют старый узел управления от цехов. Диверсанты приспособили его помещения для своей мастерской.',8,36),
 ('sleheyron','fortified_transfer_yards','Укреплённые перегрузочные дворы','Контейнерные баррикады разделяют большие площадки на узкие проезды. Банда держит здесь основные запасы и вооружённую охрану.',9,41),
 ('sleheyron','raider_refinery_hold','Опорный пункт в перерабатывающем комплексе','Уцелевшие корпуса превращены в убежище заводской банды. Боевики и переделанные рабочие машины перекрывают внутренние переходы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Toydaria.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('toydaria','marsh_landing_boards','Болотные посадочные мостки','Прочные настилы соединяют небольшую посадочную площадку с хозяйственными навесами. Грабители заняли пустующие кладовые у воды.',1,1),
 ('toydaria','algae_lake_walkways','Мостки водорослевого озера','Узкие дорожки проходят над плотными зелёными матами. Хищники подходят к опорам, а черви скрываются под илистой поверхностью.',2,6),
 ('toydaria','abandoned_harvest_sheds','Покинутые навесы сборщиков','Пустые корзины и старые сушильные рамы стоят под низкими крышами. Налётчики используют их для сортировки похищенного продовольствия.',3,11),
 ('toydaria','reed_channel_crossing','Переход над заросшей протокой','Дощатый мост соединяет две устойчивые опоры над мутной водой. Среди растительности скрываются болотные звери.',4,16),
 ('toydaria','old_water_intake','Старый водозабор','Насосные будки стоят на укреплённых площадках возле озера. Сбойные машины работают среди залежей брошенного оборудования.',5,21),
 ('toydaria','raider_store_walkways','Настилы разбойничьих кладовых','На удалённом участке хозяйства грабители соединили навесы мостками. Они охраняют украденные припасы и захваченные рабочие механизмы.',6,26),
 ('toydaria','deep_muck_lake_edge','Край глубокого илистого озера','Маршрут идёт по приподнятым мосткам вдоль тёмных заводей. Из зарослей к настилу подбираются крупные хищники.',7,31),
 ('toydaria','flooded_farm_platforms','Платформы затопленного хозяйства','От старых участков сохранились высокие настилы и водозаборные будки. Внизу поселились черви, наверху ещё движутся повреждённые машины.',8,36),
 ('toydaria','grabworm_pool_walkway','Мостки над логовищами хватателей','Укреплённый проход огибает глубокие ямы в илистом дне. Черви-хвататели охраняют свои укрытия под водорослевыми матами.',9,41),
 ('toydaria','hidden_provision_hold','Тайное хранилище припасов','Кладовые на высоких опорах скрыты за кольцом болотных зарослей. Банда защищает свои запасы и старые насосные площадки.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Kwenn.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('kwenn','surface_transfer_landing','Наземная перевалочная площадка','Грузовые настилы окружены невысокими складами. В закрытых секциях скрываются похитители контейнеров и их сообщники.',1,1),
 ('kwenn','outer_container_rows','Внешние ряды контейнеров','Старые грузовые ячейки образуют тесные проходы. Налётчики устроили в них укрытия и наблюдательные посты.',2,6),
 ('kwenn','idle_repair_sheds','Пустующие ремонтные навесы','Под металлическими крышами осталось оборудование перевозчиков. Преступные техники разбирают машины рядом со сбойными дроидами.',3,11),
 ('kwenn','service_cable_trenches','Служебные кабельные траншеи','Укреплённые проходы соединяют подземные коммуникации грузового района. Здесь проложен скрытый маршрут похитителей оборудования.',4,16),
 ('kwenn','seized_dispatch_yard','Захваченный диспетчерский двор','Группа налётчиков удерживает пустующий узел распределения грузов. Часовые прикрывают подходы к складам с краденым.',5,21),
 ('kwenn','closed_loading_halls','Закрытые погрузочные залы','Длинные крытые линии давно остановлены. Сбойные механизмы продолжают движение между опорами и пустыми приёмниками.',6,26),
 ('kwenn','hidden_machine_workshops','Тайные машинные мастерские','Перегородки скрывают цеха переделки украденной техники. Банда разместила здесь вооружённую охрану и рабочие машины.',7,31),
 ('kwenn','underground_freight_passages','Подземные грузовые проходы','Широкие тоннели идут под складским районом. Налётчики используют их для перевозки добычи между укреплёнными дворами.',8,36),
 ('kwenn','outlaw_warehouse_block','Складской блок преступной группы','Ряды закрытых ангаров защищены самодельными заграждениями. Часовые не подпускают посторонних к основному хранилищу банды.',9,41),
 ('kwenn','fortified_cargo_terminal','Укреплённый грузовой терминал','Старый наземный терминал стал опорным пунктом налётчиков. Внутренние посты прикрывают склады и подпольные ремонтные линии.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Nar Kreeta.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('nar_kreeta','cliff_supply_landing','Грузовая площадка у скалы','Небольшой посадочный настил примыкает к каменной стене. Налётчики заняли пустующие складские ниши вдоль дороги.',1,1),
 ('nar_kreeta','rocky_switchback_trail','Тропа скальных поворотов','Узкий путь поднимается между каменными выступами. Звери прячутся в расщелинах, а грабители наблюдают за дорогой сверху.',2,6),
 ('nar_kreeta','abandoned_caravan_halls','Покинутые караванные залы','В толще горы сохранились помещения старого перевалочного пункта. Банда разместила здесь украденные товары и свою охрану.',3,11),
 ('nar_kreeta','outer_cave_ledges','Уступы внешних пещер','Каменные полки соединяют широкие входы в природные полости. Хищники занимают боковые лазы и выходят на охоту вдоль тропы.',4,16),
 ('nar_kreeta','old_boring_gallery','Старая проходческая галерея','Закреплённый тоннель уходит от склона к покинутой выработке. Неисправные дроиды продолжают движение вдоль пустых грузовых полос.',5,21),
 ('nar_kreeta','raider_canyon_camp','Лагерь налётчиков в каньоне','Скальные стены скрывают навесы и запасы преступной группы. Боевики охраняют переход к подземным кладовым.',6,26),
 ('nar_kreeta','deep_predator_hollows','Глубокие логовища в скалах','Цепочка природных залов отделена от старых выработок узкими лазами. Здесь укрываются самые крупные пещерные звери.',7,31),
 ('nar_kreeta','sealed_service_tunnels','Закрытые служебные тоннели','Банда пробила проходы между покинутыми рабочими секторами. Вооружённые дозоры движутся рядом с переделанными машинами.',8,36),
 ('nar_kreeta','fortified_rock_storehouses','Укреплённые скальные склады','Тяжёлые перегородки защищают хранилища краденых грузов. Боевики занимают узкие переходы между каменными залами.',9,41),
 ('nar_kreeta','inner_outlaw_galleries','Внутренние галереи банды','Глубокий складской комплекс превращён в убежище налётчиков. Вооружённая охрана и повреждённые машины прикрывают его внутренние помещения.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Klatooine.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('klatooine','outer_caravan_landing','Окраинная караванная площадка','Посадочный настил стоит рядом с пустующими дорожными навесами. Грабители хранят в них товары, отнятые у перевозчиков.',1,1),
 ('klatooine','dry_ridge_track','Дорога сухих гряд','Каменные возвышения разбивают песчаную равнину на узкие ложбины. Среди уступов прячутся хищники и наблюдатели банды.',2,6),
 ('klatooine','abandoned_waystation','Заброшенный дорожный пост','Старые кладовые окружают крытый двор для грузов. Налётчики заняли пустые помещения и выставили вооружённые посты.',3,11),
 ('klatooine','shaded_sand_hollows','Тенистые песчаные ложбины','Под нависающими скалами сохраняются прохлада и глубокие норы. Звери защищают укрытия вдоль извилистой тропы.',4,16),
 ('klatooine','old_water_station','Старая водозаборная станция','В заброшенном хозяйственном дворе сохранились насосы и закрытые ёмкости. Сбойные дроиды всё ещё обходят площадки обслуживания.',5,21),
 ('klatooine','raider_caravan_yard','Караванный двор налётчиков','Украденные контейнеры сложены между низкими стенами. Вооружённая группа готовит здесь снаряжение для новых нападений.',6,26),
 ('klatooine','wind_worn_stone_pass','Проход выветренных скал','Сухие каменные арки соединяют несколько укрытых от ветра ложбин. Местные хищники используют боковые ниши как логовища.',7,31),
 ('klatooine','buried_supply_vaults','Занесённые песком хранилища','Под старым двором сохранились прочные грузовые помещения. Банда расчистила входы и использует уцелевшие машины для охраны.',8,36),
 ('klatooine','fortified_well_compound','Укреплённый двор у скважины','Высокая ограда окружает захваченный пункт снабжения. Охрана банды удерживает склады и насосные площадки.',9,41),
 ('klatooine','remote_outlaw_hold','Дальнее убежище караванной банды','Скрытый за грядой комплекс связан с пустынными дорогами узкими проходами. Банда охраняет основные запасы похищенных товаров.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Vodran.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('vodran','raised_river_landing','Приподнятая речная площадка','Высокий настил соединяет посадочный участок с причалом. Грабители заняли пустые грузовые навесы у края джунглей.',1,1),
 ('vodran','root_bank_trail','Корневая тропа вдоль берега','Путь проходит по плотным переплетениям корней над топкой землёй. Хищники наблюдают из зарослей, а в воде скрываются дианоги.',2,6),
 ('vodran','abandoned_supply_piers','Покинутые причалы снабжения','Крепкие опоры удерживают старые складские платформы над протокой. Налётчики используют их вместе с оставшейся рабочей техникой.',3,11),
 ('vodran','shallow_bayou_boards','Мостки мелкой заводи','Настилы огибают илистые ямы и низкие островки. Дианоги укрываются под корнями возле самых узких переходов.',4,16),
 ('vodran','dense_jungle_ridge','Гряда густых джунглей','Небольшое возвышение проходит между заболоченными низинами. Крупные звери занимают сухие ниши среди толстых корней.',5,21),
 ('vodran','hidden_river_store','Тайный речной склад','Грабители укрепили заброшенную платформу в стороне от основного русла. Под навесами сложены похищенные припасы и механизмы.',6,26),
 ('vodran','dianoga_channel_bank','Берег протоки дианог','Укреплённая дорожка тянется вдоль тёмной воды. В глубоких карманах под берегом поселились крупные дианоги.',7,31),
 ('vodran','overgrown_work_platforms','Заросшие рабочие платформы','Старые площадки обслуживания соединены высокими мостками. Повреждённые машины движутся над водой, где укрываются болотные существа.',8,36),
 ('vodran','deep_root_den_paths','Тропы глубоких корневых логовищ','Густые переплетения корней образуют сухие переходы над болотом. Джунглевые звери охраняют ниши, а дианоги скрывают выходы из воды.',9,41),
 ('vodran','fortified_raider_piers','Укреплённые причалы налётчиков','Грузовые платформы преступной группы скрыты за плотными зарослями. Банда удерживает склады и подходы к ремонтным настилам.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Kintan.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('kintan','rockland_supply_landing','Грузовая площадка каменных равнин','Плоский настил примыкает к старому двору снабжения. Похитители руды заняли навесы рядом с брошенной карьерной техникой.',1,1),
 ('kintan','weathered_wasteland_track','Тропа выветренных пустошей','Дорога огибает сухие скалы и широкие каменные ложбины. Хищники занимают тенистые ниши вдоль перехода.',2,6),
 ('kintan','old_quarry_yards','Дворы старого карьера','На остановленном участке сохранились грузовые рампы и ремонтные навесы. Налётчики собирают оборудование, пока сбойные машины продолжают обходы.',3,11),
 ('kintan','wooded_valley_path','Тропа лесистой долины','За сухими возвышениями начинается защищённый лесной участок. Звери прячутся среди корней и охраняют подходы к своим укрытиям.',4,16),
 ('kintan','marsh_edge_causeway','Переход вдоль болотной кромки','Каменная насыпь отделяет тропу от влажных низин. Из густой растительности к дороге выходят лесные хищники.',5,21),
 ('kintan','raider_ore_store','Рудный склад налётчиков','Заброшенные кладовые скрыты в складках каменного рельефа. Грабители охраняют добычу и захваченные рабочие машины.',6,26),
 ('kintan','archipelago_shore_trail','Береговая тропа архипелага','Путь проходит по суше вдоль защищённых бухт и лесистых склонов. Хищники держатся у густых зарослей рядом с берегом.',7,31),
 ('kintan','cooled_lava_ridge','Гребень застывшей лавы','Устойчивая каменная тропа идёт по давно остывшему участку вулканического района. Звери занимают расщелины, защищённые от горячих ветров.',8,36),
 ('kintan','deep_boring_works','Глубокие проходческие выработки','Закреплённые галереи соединяют закрытые участки добычи. Повреждённые дроиды продолжают работу среди тайников похитителей руды.',9,41),
 ('kintan','fortified_mining_hold','Укреплённый пункт у выработок','Банда превратила старые служебные корпуса в защищённый склад. Вооружённые посты и переделанные машины перекрывают подходы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Ubrikkia.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('ubrikkia','industrial_freight_landing','Промышленная грузовая площадка','Посадочный настил выходит к старым складам деталей. Налётчики заняли пустые секции рядом с брошенными разборочными машинами.',1,1),
 ('ubrikkia','outer_scrap_lanes','Проезды окраинного разборочного двора','Ряды снятых корпусов образуют узкие проходы. Похитители деталей скрываются среди оборудования и повреждённых дроидов.',2,6),
 ('ubrikkia','idle_assembly_workshop','Остановленная сборочная мастерская','Крытая линия давно лишилась поставок и персонала. Сбойные машины продолжают движение между пустыми рабочими местами.',3,11),
 ('ubrikkia','parts_warehouse_rows','Ряды складов комплектующих','Закрытые ангары соединены погрузочными галереями. Банда использует уцелевшие помещения для сортировки похищенных деталей.',4,16),
 ('ubrikkia','service_conduit_passages','Проходы служебных коммуникаций','Тоннели под мастерскими сохранили настилы и кабельные опоры. Преступники используют их для скрытой перевозки оборудования.',5,21),
 ('ubrikkia','illegal_repair_yards','Подпольные ремонтные дворы','За глухой оградой переделывают украденные механизмы. Вооружённая охрана прикрывает рабочие площадки со сбойными дроидами.',6,26),
 ('ubrikkia','closed_reclamation_halls','Закрытые залы разборки','Стеллажи с корпусами тянутся вдоль остановленных конвейеров. Повреждённые машины пытаются продолжать работу без управления.',7,31),
 ('ubrikkia','seized_machine_vaults','Захваченные машинные хранилища','Укреплённые помещения заполнены краденым оборудованием. Боевики охраняют проходы, поддерживая работу уцелевших сборочных машин.',8,36),
 ('ubrikkia','automated_loading_block','Блок автоматической погрузки','Крытые грузовые полосы соединяют дальние цеха с закрытым двором. Сбойные механизмы продолжают перемещаться среди пустых контейнеров.',9,41),
 ('ubrikkia','fortified_outlaw_factory','Укреплённый заводской корпус банды','Преступная группа заняла внутренние мастерские и главные склады. Боевики и повреждённые машины защищают её запасы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('sleheyron','outer_freight_landing','sleheyron_dock_raider','sleheyron_cargo_saboteur','sleheyron_dock_raider'),
 ('sleheyron','basalt_service_road','sleheyron_dock_raider','sleheyron_cargo_saboteur','sleheyron_gang_enforcer'),
 ('sleheyron','idle_processing_yard','sleheyron_cargo_saboteur','sleheyron_furnace_droid','sleheyron_furnace_droid'),
 ('sleheyron','pipe_access_galleries','sleheyron_cargo_saboteur','sleheyron_furnace_droid','sleheyron_cargo_saboteur'),
 ('sleheyron','stolen_fuel_depot','sleheyron_dock_raider','sleheyron_gang_enforcer','sleheyron_gang_enforcer'),
 ('sleheyron','sealed_furnace_halls','sleheyron_furnace_droid','sleheyron_furnace_droid','sleheyron_furnace_droid'),
 ('sleheyron','slag_ridge_passage','sleheyron_dock_raider','sleheyron_gang_enforcer','sleheyron_cargo_saboteur'),
 ('sleheyron','abandoned_dispatch_block','sleheyron_cargo_saboteur','sleheyron_furnace_droid','sleheyron_gang_enforcer'),
 ('sleheyron','fortified_transfer_yards','sleheyron_dock_raider','sleheyron_gang_enforcer','sleheyron_gang_enforcer'),
 ('sleheyron','raider_refinery_hold','sleheyron_gang_enforcer','sleheyron_cargo_saboteur','sleheyron_furnace_droid'),
 ('toydaria','marsh_landing_boards','toydaria_harvest_raider','toydaria_pump_droid','toydaria_harvest_raider'),
 ('toydaria','algae_lake_walkways','toydaria_marsh_predator','toydaria_grabworm','toydaria_marsh_predator'),
 ('toydaria','abandoned_harvest_sheds','toydaria_harvest_raider','toydaria_harvest_raider','toydaria_pump_droid'),
 ('toydaria','reed_channel_crossing','toydaria_marsh_predator','toydaria_grabworm','toydaria_grabworm'),
 ('toydaria','old_water_intake','toydaria_pump_droid','toydaria_pump_droid','toydaria_pump_droid'),
 ('toydaria','raider_store_walkways','toydaria_harvest_raider','toydaria_pump_droid','toydaria_harvest_raider'),
 ('toydaria','deep_muck_lake_edge','toydaria_marsh_predator','toydaria_grabworm','toydaria_grabworm'),
 ('toydaria','flooded_farm_platforms','toydaria_pump_droid','toydaria_grabworm','toydaria_pump_droid'),
 ('toydaria','grabworm_pool_walkway','toydaria_grabworm','toydaria_grabworm','toydaria_grabworm'),
 ('toydaria','hidden_provision_hold','toydaria_harvest_raider','toydaria_pump_droid','toydaria_harvest_raider'),
 ('kwenn','surface_transfer_landing','kwenn_cargo_thief','kwenn_raider_technician','kwenn_cargo_thief'),
 ('kwenn','outer_container_rows','kwenn_cargo_thief','kwenn_outlaw_sentry','kwenn_cargo_thief'),
 ('kwenn','idle_repair_sheds','kwenn_raider_technician','kwenn_service_droid','kwenn_raider_technician'),
 ('kwenn','service_cable_trenches','kwenn_cargo_thief','kwenn_service_droid','kwenn_outlaw_sentry'),
 ('kwenn','seized_dispatch_yard','kwenn_cargo_thief','kwenn_outlaw_sentry','kwenn_raider_technician'),
 ('kwenn','closed_loading_halls','kwenn_service_droid','kwenn_service_droid','kwenn_service_droid'),
 ('kwenn','hidden_machine_workshops','kwenn_raider_technician','kwenn_service_droid','kwenn_outlaw_sentry'),
 ('kwenn','underground_freight_passages','kwenn_cargo_thief','kwenn_outlaw_sentry','kwenn_service_droid'),
 ('kwenn','outlaw_warehouse_block','kwenn_outlaw_sentry','kwenn_raider_technician','kwenn_outlaw_sentry'),
 ('kwenn','fortified_cargo_terminal','kwenn_outlaw_sentry','kwenn_service_droid','kwenn_raider_technician'),
 ('nar_kreeta','cliff_supply_landing','nar_kreeta_cargo_raider','nar_kreeta_mining_droid','nar_kreeta_cargo_raider'),
 ('nar_kreeta','rocky_switchback_trail','nar_kreeta_cargo_raider','nar_kreeta_tunnel_stalker','nar_kreeta_cargo_raider'),
 ('nar_kreeta','abandoned_caravan_halls','nar_kreeta_cargo_raider','nar_kreeta_outlaw_enforcer','nar_kreeta_outlaw_enforcer'),
 ('nar_kreeta','outer_cave_ledges','nar_kreeta_tunnel_stalker','nar_kreeta_tunnel_stalker','nar_kreeta_tunnel_stalker'),
 ('nar_kreeta','old_boring_gallery','nar_kreeta_mining_droid','nar_kreeta_mining_droid','nar_kreeta_mining_droid'),
 ('nar_kreeta','raider_canyon_camp','nar_kreeta_cargo_raider','nar_kreeta_outlaw_enforcer','nar_kreeta_outlaw_enforcer'),
 ('nar_kreeta','deep_predator_hollows','nar_kreeta_tunnel_stalker','nar_kreeta_tunnel_stalker','nar_kreeta_tunnel_stalker'),
 ('nar_kreeta','sealed_service_tunnels','nar_kreeta_cargo_raider','nar_kreeta_mining_droid','nar_kreeta_outlaw_enforcer'),
 ('nar_kreeta','fortified_rock_storehouses','nar_kreeta_outlaw_enforcer','nar_kreeta_cargo_raider','nar_kreeta_outlaw_enforcer'),
 ('nar_kreeta','inner_outlaw_galleries','nar_kreeta_outlaw_enforcer','nar_kreeta_mining_droid','nar_kreeta_outlaw_enforcer'),
 ('klatooine','outer_caravan_landing','klatooine_caravan_raider','klatooine_pump_droid','klatooine_caravan_raider'),
 ('klatooine','dry_ridge_track','klatooine_dune_stalker','klatooine_caravan_raider','klatooine_dune_stalker'),
 ('klatooine','abandoned_waystation','klatooine_caravan_raider','klatooine_raider_guard','klatooine_raider_guard'),
 ('klatooine','shaded_sand_hollows','klatooine_dune_stalker','klatooine_dune_stalker','klatooine_dune_stalker'),
 ('klatooine','old_water_station','klatooine_pump_droid','klatooine_pump_droid','klatooine_pump_droid'),
 ('klatooine','raider_caravan_yard','klatooine_caravan_raider','klatooine_raider_guard','klatooine_raider_guard'),
 ('klatooine','wind_worn_stone_pass','klatooine_dune_stalker','klatooine_dune_stalker','klatooine_dune_stalker'),
 ('klatooine','buried_supply_vaults','klatooine_caravan_raider','klatooine_pump_droid','klatooine_raider_guard'),
 ('klatooine','fortified_well_compound','klatooine_raider_guard','klatooine_pump_droid','klatooine_raider_guard'),
 ('klatooine','remote_outlaw_hold','klatooine_caravan_raider','klatooine_raider_guard','klatooine_raider_guard'),
 ('vodran','raised_river_landing','vodran_river_raider','vodran_salvage_droid','vodran_river_raider'),
 ('vodran','root_bank_trail','vodran_jungle_stalker','vodran_dianoga','vodran_jungle_stalker'),
 ('vodran','abandoned_supply_piers','vodran_river_raider','vodran_salvage_droid','vodran_salvage_droid'),
 ('vodran','shallow_bayou_boards','vodran_dianoga','vodran_dianoga','vodran_dianoga'),
 ('vodran','dense_jungle_ridge','vodran_jungle_stalker','vodran_jungle_stalker','vodran_jungle_stalker'),
 ('vodran','hidden_river_store','vodran_river_raider','vodran_salvage_droid','vodran_river_raider'),
 ('vodran','dianoga_channel_bank','vodran_dianoga','vodran_dianoga','vodran_dianoga'),
 ('vodran','overgrown_work_platforms','vodran_salvage_droid','vodran_dianoga','vodran_salvage_droid'),
 ('vodran','deep_root_den_paths','vodran_jungle_stalker','vodran_dianoga','vodran_jungle_stalker'),
 ('vodran','fortified_raider_piers','vodran_river_raider','vodran_salvage_droid','vodran_river_raider'),
 ('kintan','rockland_supply_landing','kintan_ore_raider','kintan_quarry_droid','kintan_ore_raider'),
 ('kintan','weathered_wasteland_track','kintan_wasteland_predator','kintan_wasteland_predator','kintan_wasteland_predator'),
 ('kintan','old_quarry_yards','kintan_ore_raider','kintan_quarry_droid','kintan_quarry_droid'),
 ('kintan','wooded_valley_path','kintan_forest_stalker','kintan_forest_stalker','kintan_forest_stalker'),
 ('kintan','marsh_edge_causeway','kintan_forest_stalker','kintan_forest_stalker','kintan_forest_stalker'),
 ('kintan','raider_ore_store','kintan_ore_raider','kintan_quarry_droid','kintan_ore_raider'),
 ('kintan','archipelago_shore_trail','kintan_forest_stalker','kintan_forest_stalker','kintan_forest_stalker'),
 ('kintan','cooled_lava_ridge','kintan_wasteland_predator','kintan_wasteland_predator','kintan_wasteland_predator'),
 ('kintan','deep_boring_works','kintan_quarry_droid','kintan_ore_raider','kintan_quarry_droid'),
 ('kintan','fortified_mining_hold','kintan_ore_raider','kintan_quarry_droid','kintan_ore_raider'),
 ('ubrikkia','industrial_freight_landing','ubrikkia_cargo_raider','ubrikkia_salvage_droid','ubrikkia_cargo_raider'),
 ('ubrikkia','outer_scrap_lanes','ubrikkia_cargo_raider','ubrikkia_salvage_droid','ubrikkia_salvage_droid'),
 ('ubrikkia','idle_assembly_workshop','ubrikkia_assembly_droid','ubrikkia_assembly_droid','ubrikkia_assembly_droid'),
 ('ubrikkia','parts_warehouse_rows','ubrikkia_cargo_raider','ubrikkia_illegal_guard','ubrikkia_cargo_raider'),
 ('ubrikkia','service_conduit_passages','ubrikkia_cargo_raider','ubrikkia_salvage_droid','ubrikkia_illegal_guard'),
 ('ubrikkia','illegal_repair_yards','ubrikkia_illegal_guard','ubrikkia_assembly_droid','ubrikkia_illegal_guard'),
 ('ubrikkia','closed_reclamation_halls','ubrikkia_salvage_droid','ubrikkia_salvage_droid','ubrikkia_salvage_droid'),
 ('ubrikkia','seized_machine_vaults','ubrikkia_illegal_guard','ubrikkia_assembly_droid','ubrikkia_illegal_guard'),
 ('ubrikkia','automated_loading_block','ubrikkia_assembly_droid','ubrikkia_salvage_droid','ubrikkia_assembly_droid'),
 ('ubrikkia','fortified_outlaw_factory','ubrikkia_illegal_guard','ubrikkia_assembly_droid','ubrikkia_salvage_droid')
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
WHERE p.code IN ('sleheyron','toydaria','kwenn','nar_kreeta','klatooine','vodran','kintan','ubrikkia') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
