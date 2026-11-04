-- Second additional protected-core batch; membership/control remain in seed07.
-- Districts and enemy roles are authored around the worlds' geography.
-- Early Great Galactic War: no later radioactive devastation of Duro, Clone Wars
-- occupation of Brentaal or Vader-era events on Jaguada and its separate moon.
-- Khar Delba's planetary citadel is distinct from the hidden base on Khar Shian.
-- Chandrila's peaceful wildlife is not a combat target; Rhelg retains its farmland.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('chandrila_crop_raider','Налётчик сельских дорог','Грабит небольшие обозы с урожаем Чандрилы. Прячет добычу в пустующих амбарах и требует плату за проезд по удалённым дорогам.','humanoid'),
 ('chandrila_wildlife_smuggler','Контрабандист редких животных','Незаконно отлавливает охраняемых животных для подпольных покупателей. Защищает оружием ловушки, перевалочные стоянки и клетки с добычей.','humanoid'),
 ('chandrila_harvest_droid','Неисправный уборочный дроид','Сельскохозяйственная машина утратила ограничения безопасности. Продолжает обрабатывать заброшенные участки, не различая растения и оказавшихся поблизости путников.','droid'),
 ('chandrila_estate_sentry','Взломанный сторож усадьбы','Бывший охранный дроид служит преступникам, занявшим пустующую усадьбу. Перекрывает подходы к тайникам и игнорирует прежние пропуска.','droid'),
 ('duro_salvage_raider','Налётчик промышленных окраин','Присваивает пригодные детали из старых комплексов Дуро. Нападает на ремонтные бригады и удерживает нелегальные склады металла.','humanoid'),
 ('duro_contraband_technician','Техник подпольных мастерских','Переделывает краденые промышленные узлы для контрабандистов. Вооружён и готов защищать мастерскую вместе с перепрограммированной автоматикой.','humanoid'),
 ('duro_reclamation_droid','Сбойный восстановительный дроид','Разбирал повреждённые конструкции на промышленной площадке. После сбоя воспринимает любой подвижный объект как материал для переработки.','droid'),
 ('duro_industrial_warden','Захваченный промышленный страж','Тяжёлая охранная машина получила чужие коды управления. Контролирует присвоенные производственные помещения и атакует посетителей без допуска преступной группы.','droid'),
 ('brentaal_freight_raider','Налётчик торговых маршрутов','Охотится за грузами на складских окраинах Брентаала IV. Использует горные объезды и пустующие терминалы, чтобы скрыться от законных патрулей.','humanoid'),
 ('brentaal_customs_forger','Вооружённый подделыватель допусков','Снабжает контрабандистов ложными грузовыми документами. Охраняет тайные архивы и устраняет свидетелей незаконных перегрузок.','humanoid'),
 ('brentaal_cargo_automaton','Сбойный грузовой автомат','Старая погрузочная машина действует по повреждённой программе. Преграждает проходы между контейнерами и применяет захваты против приближающихся существ.','droid'),
 ('brentaal_kundril','Прибрежный кундрил','Крупный летающий хищник Брентаала IV находит добычу по звуку. Появляется у морских окраин и опасен для тех, кто потревожит его кормовой участок.','beast'),
 ('jaguada_dune_stalker','Хищник дюн Джагуады','Прячется в тени каменных гряд среди пустыни. Выслеживает добычу у заброшенных колодцев и укрытых от ветра проходов.','beast'),
 ('jaguada_relic_smuggler','Контрабандист реликвий Джагуады','Вывозит находки из древних сооружений без разрешения властей. Держит вооружённую охрану у самовольных раскопок и тайных складов.','humanoid'),
 ('jaguada_excavation_droid','Повреждённый раскопочный дроид','Освобождал древние стены Джагуады от песка и обломков. Сломанные датчики заставляют его применять рабочие инструменты против любых движущихся помех.','droid'),
 ('jaguada_renegade_adept','Отступник пустынных святилищ','Изучает запретные реликвии вдали от имперского надзора. Удерживает древние залы с помощью вооружённых сообщников и захваченной техники.','humanoid'),
 ('rhelg_field_stalker','Хищник заброшенных угодий','Обосновался среди одичавших полей и заросших межей Релга. Подбирается к добыче по осушительным канавам и покинутым хозяйственным дворам.','beast'),
 ('rhelg_supply_raider','Расхититель сельских запасов','Грабит перевозчиков и старые зерновые склады. Защищает награбленное в пустующих постройках, откуда удобно следить за дорогами.','humanoid'),
 ('rhelg_harvest_droid','Сбойный полевой дроид','Земледельческий автомат давно лишён надёжного обслуживания. Его режущие инструменты продолжают работать даже при появлении людей на участке.','droid'),
 ('rhelg_palace_guardian','Повреждённый дворцовый страж','Старая охранная машина патрулирует закрытые сооружения Релга. Утраченные ключи опознавания превратили её в угрозу для любого посетителя.','droid'),
 ('khar_delba_ice_stalker','Хищник ледяных ложбин','Скрывается от ветра между снежными наметами и каменными уступами Кхар-Дельбы. Подстерегает добычу у пещер и защищённых горных троп.','beast'),
 ('khar_delba_relic_raider','Расхититель ледяных руин','Ищет ценности в промёрзших сооружениях Кхар-Дельбы. Вооружён и охраняет найденные проходы, не допуская туда соперников и свидетелей.','humanoid'),
 ('khar_delba_sentry_droid','Промёрзший механический часовой','Древняя охранная машина сохранила подвижность под ледяной коркой. Повреждённые цепи управления заставляют её атаковать всех, кто входит в охраняемый участок.','droid'),
 ('khar_delba_war_behemoth','Ситхский боевой исполин','Огромный зверь Кхар-Дельбы, подобных которому древние ситхи использовали в войнах. Опасные особи обитают вдали от обжитых площадок, среди кратеров и старых укреплений.','beast')
ON CONFLICT (code) DO NOTHING;

-- Republic core: Chandrila, Duro and Brentaal IV. Duro districts are on the surface.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('chandrila','rural_landing','Сельская посадочная площадка','Небольшой терминал окружён полями и дорогами к фермерским поселениям. У дальних погрузочных навесов скрываются похитители урожая.',1,1),
 ('chandrila','forest_margins','Лесные межи','Лесные полосы отделяют возделанные участки от охраняемых природных земель. Контрабандисты прячут среди деревьев ловушки и стоянки для вывоза животных.',2,6),
 ('chandrila','silver_sea_coves','Бухты Серебряного моря','За тихими пляжами тянутся закрытые причалы небольших хозяйств. Преступники используют пустующие пристани для тайных перегрузок.',3,11),
 ('chandrila','orchard_storehouses','Склады плодовых садов','Ряды амбаров стоят между старыми садами и грузовой дорогой. Часть построек занята налётчиками, а в заброшенных проходах действует неисправная техника.',4,16),
 ('chandrila','crystal_canyon_trails','Тропы Кристальных каньонов','Свет отражается от граней скальных стен над извилистыми тропами. В боковых ущельях контрабандисты устроили укрытия и охраняемые перевалочные пункты.',5,21),
 ('chandrila','irrigation_terraces','Оросительные террасы','Каналы соединяют ступенчатые поля с хозяйственными площадками. На закрытых участках брошенные автоматы продолжают опасную работу без надзора.',6,26),
 ('chandrila','illegal_capture_compound','Подпольный отловочный двор','Среди деревьев спрятаны клетки, грузовые платформы и замаскированные навесы. Торговцы редкими животными выставили вооружённую охрану у каждого выхода.',7,31),
 ('chandrila','old_country_estate','Старая загородная усадьба','Заросшие аллеи ведут к пустующим жилым корпусам. Преступная группа восстановила охранных дроидов и превратила усадьбу в склад награбленного.',8,36),
 ('chandrila','closed_agricultural_station','Закрытая агростанция','За ограждениями сохранились опытные поля и помещения сельскохозяйственной автоматики. Нелегальные хозяева испытывают здесь переделанные машины и охраняют технические архивы.',9,41),
 ('chandrila','hidden_manor','Скрытое поместье','Глухая лесная дорога заканчивается у старых ворот. За ними находится укреплённое убежище организаторов сельских налётов и торговли охраняемыми животными.',10,46),
 ('duro','surface_freight_yard','Наземный грузовой двор','Грузовые опоры стоят у действующих служебных переходов промышленного района. Поодаль от рабочих маршрутов налётчики разбирают оставленные контейнеры.',1,1),
 ('duro','restoration_blocks','Восстановительные кварталы','Старые повреждённые корпуса соседствуют с укреплёнными проходами ремонтных бригад. На закрытых площадках осталась техника с неисправными датчиками.',2,6),
 ('duro','scrap_sorting_fields','Площадки сортировки металла','Кучи разобранных конструкций разделены узкими проездами. Похитители деталей держат здесь тайники и мешают законным работам по восстановлению.',3,11),
 ('duro','old_power_corridors','Старые силовые коридоры','Закрытые технические переходы связывают производственные блоки. Подпольные мастера подключили свои цеха к уцелевшим линиям питания.',4,16),
 ('duro','sealed_processing_halls','Закрытые перерабатывающие цеха','Обслуживаемые галереи проходят над остановленными производственными линиями. На нижних площадках действуют сбойные автоматы и охрана нелегальных разборщиков.',5,21),
 ('duro','reclaimed_factory_yards','Расчищенные заводские дворы','После разборки старых завалов открылись проезды к забытым складам. Вооружённые группы пытаются удержать найденное оборудование и восстановительные машины.',6,26),
 ('duro','underground_workshops','Подпольные ремонтные мастерские','В глубине пустующих корпусов снова горит рабочий свет. Здесь переделывают краденые узлы и собирают охранных дроидов для преступных заказчиков.',7,31),
 ('duro','industrial_record_vault','Хранилище промышленных схем','Укреплённые архивные помещения соединены закрытыми переходами. Взломщики охраняют захваченные чертежи и коды управления заводским оборудованием.',8,36),
 ('duro','automated_security_depot','База охранной автоматики','Ряды сервисных боксов окружают испытательную площадку. Незаконные хозяева вернули в строй тяжёлых стражей, лишив их прежних ограничений.',9,41),
 ('duro','seized_production_complex','Захваченный производственный комплекс','Группа контрабандистов перекрыла ворота большого наземного завода. За ограждениями сосредоточены украденные машины, переделанная охрана и запасы ценных деталей.',10,46),
 ('brentaal_iv','traders_landing','Торговая посадочная площадка','Погрузочные дороги сходятся у края оживлённого терминала. В стороне от основных потоков налётчики высматривают плохо охраняемые партии товара.',1,1),
 ('brentaal_iv','valley_warehouse_rows','Складские ряды долины','Длинные здания занимают ровные участки между каменными склонами. Пустующие секции служат укрытиями для грузовых воров и сбойной погрузочной техники.',2,6),
 ('brentaal_iv','coastal_freight_road','Прибрежная грузовая дорога','Транспортный путь проходит между морем и складскими окраинами. У покинутых съездов скрываются налётчики, а шум у воды привлекает кундрилов.',3,11),
 ('brentaal_iv','canyon_transit_yard','Перевалочный двор в каньоне','Высокие скальные стены укрывают старые погрузочные площадки. Контрабандисты подменяют здесь документы и перегружают похищенные контейнеры.',4,16),
 ('brentaal_iv','mountain_warehouse_steps','Горные складские террасы','Грузовые подъёмники соединяют площадки на выровненных уступах. Заброшенные склады заняты вооружёнными посредниками и неисправными автоматами.',5,21),
 ('brentaal_iv','abandoned_sea_terminal','Покинутый морской терминал','Старые причалы выходят к открытой воде за городской окраиной. Рядом кормятся кундрилы, а крытые ангары стали тайниками похитителей грузов.',6,26),
 ('brentaal_iv','forged_manifest_archive','Архив поддельных ведомостей','За обычным складским фасадом скрыта мастерская грузовых допусков. Её хозяева охраняют записи о контрабандных поставках и украденных партиях.',7,31),
 ('brentaal_iv','closed_plateau_port','Закрытый порт на плато','Выровненная горная площадка хранит ряды неиспользуемых ангаров. Преступники восстановили погрузочные линии и перекрыли наземные подходы.',8,36),
 ('brentaal_iv','seacliff_loading_galleries','Погрузочные галереи морских скал','Крытые эстакады соединяют склады над обрывистым берегом. У внешних площадок охотятся кундрилы, внутри действуют незаконные перевозчики и сбойная техника.',9,41),
 ('brentaal_iv','smugglers_distribution_hub','Распределительный узел контрабандистов','Закрытые складские дворы образуют укреплённый перегрузочный комплекс. Здесь сходятся похищенные грузы, поддельные документы и вооружённая охрана подпольных торговцев.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Imperial core: Jaguada, Rhelg and Khar Delba; their ancient rulers are not living NPCs.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('jaguada','desert_supply_landing','Пустынная площадка снабжения','Низкие навесы защищают грузы от песка и сухого ветра. У дальних складов скрываются похитители экспедиционных запасов.',1,1),
 ('jaguada','dune_tracks','Дюнные тропы','Старые пути исчезают между песчаными грядами. В укрытых впадинах охотятся звери, а контрабандисты обходят по ним основные дороги.',2,6),
 ('jaguada','abandoned_wells','Покинутые колодцы','Каменные ограды отмечают бывшие стоянки пустынных караванов. Оставшиеся укрытия заняты хищниками и вооружёнными искателями древностей.',3,11),
 ('jaguada','buried_outer_walls','Занесённые внешние стены','Из песка выступают укрепления древнего поселения ситхов. Незаконные раскопщики разбирают уцелевшие кладовые и охраняют проходы между стенами.',4,16),
 ('jaguada','dry_excavation_trenches','Сухие раскопочные траншеи','Глубокие выемки открывают основания старых сооружений. Повреждённые рабочие дроиды мешают расхитителям, а те не допускают сюда посторонних.',5,21),
 ('jaguada','caravan_hideout','Караванное убежище','Среди скал спрятаны грузовые платформы и навесы тайной стоянки. Отсюда контрабандисты вывозят реликвии, добытые в удалённых раскопах.',6,26),
 ('jaguada','sandstone_galleries','Галереи песчаного камня','Крытые переходы древнего комплекса частично уцелели под песком. Отступники используют их для хранения находок и охраны запретных исследований.',7,31),
 ('jaguada','sealed_relic_chambers','Закрытые палаты реликвий','Тяжёлые плиты скрывают залы с остатками древних хранилищ. Самовольные исследователи восстановили проходы и держат здесь вооружённых сообщников.',8,36),
 ('jaguada','broken_sanctuary_steps','Разбитые ступени святилища','Каменные лестницы поднимаются к полуразрушенному сооружению над пустыней. Подступы заняты отступниками и переделанной раскопочной техникой.',9,41),
 ('jaguada','hidden_desert_enclave','Тайное пустынное убежище','В глубине древних стен устроен лагерь искателей запретных знаний. Его хозяева защищают собранные реликвии и скрывают следы своего присутствия.',10,46),
 ('rhelg','farmland_landing','Посадочная площадка среди полей','Дороги расходятся от небольшого терминала к сельским владениям. На покинутых хозяйственных дворах налётчики высматривают обозы с припасами.',1,1),
 ('rhelg','overgrown_field_boundaries','Заросшие межи','Густая трава скрывает границы давно оставленных участков. По канавам между полями перемещаются хищники и расхитители сельских запасов.',2,6),
 ('rhelg','old_granaries','Старые зернохранилища','Каменные амбары стоят у заброшенного грузового пути. Налётчики заняли уцелевшие помещения, а вокруг продолжает двигаться неисправная техника.',3,11),
 ('rhelg','drainage_channels','Осушительные каналы','Разветвлённые канавы проходят между полями и старыми насосными дворами. В зарослях скрываются звери, а закрытые площадки заняты сбойными автоматами.',4,16),
 ('rhelg','abandoned_farmworks','Покинутые земледельческие мастерские','Под широкими навесами сохранились ремонтные стойки и запасы инструментов. Нелегальные разборщики охраняют технику и пытаются подчинить себе полевых дроидов.',5,21),
 ('rhelg','estate_supply_road','Дорога старых владений','Мощёный путь ведёт от сельских складов к древним дворцовым сооружениям. Вооружённые группы устроили заставы у уцелевших ворот и хозяйственных пристроек.',6,26),
 ('rhelg','overgrown_estate_courts','Заросшие усадебные дворы','Растительность заполнила промежутки между каменными корпусами. Среди пустых построек охотятся звери, а старые охранные машины всё ещё обходят закрытые участки.',7,31),
 ('rhelg','kressh_palace_approach','Подступы ко дворцу Крессша','Древняя дорога заканчивается у внешних сооружений бывшего дворца Людо Крессша. Расхитители ищут безопасные проходы между повреждёнными механическими стражами.',8,36),
 ('rhelg','palace_service_passages','Дворцовые служебные переходы','Узкие коридоры соединяют хозяйственные корпуса с внутренними дворами. Проникшие сюда грабители охраняют находки, избегая действующих участков старой защиты.',9,41),
 ('rhelg','sealed_palace_wing','Закрытое дворцовое крыло','За массивными дверями сохранились залы древней резиденции. Расхитители удерживают вскрытые помещения, но глубже по-прежнему патрулируют неисправные стражи.',10,46),
 ('khar_delba','sheltered_ice_landing','Укрытая ледяная площадка','Скальные стенки защищают небольшой терминал от снежных порывов. За пределами расчищенных проходов видны следы зверей и нелегальных экспедиций.',1,1),
 ('khar_delba','snow_hollows','Снежные ложбины','Ветер наметает снег между невысокими каменными грядами. Хищники используют эти укрытия для засад у старых экспедиционных троп.',2,6),
 ('khar_delba','frozen_crater_edge','Край замёрзшего кратера','Неровный обод кратера окружён обледеневшими уступами. В скальных нишах прячутся звери и оставлены тайники искателей древних ценностей.',3,11),
 ('khar_delba','abandoned_expedition_camp','Покинутый экспедиционный лагерь','Снег занёс жилые модули и грузовые платформы старой стоянки. Расхитители вернулись за оборудованием и восстановили часть охранной автоматики.',4,16),
 ('khar_delba','icebound_caverns','Обледеневшие пещеры','Под горными склонами сохранились глубокие естественные полости. В боковых ходах обитают хищники, а у входов устроены укрытия расхитителей.',5,21),
 ('khar_delba','buried_watchworks','Занесённые сторожевые сооружения','Каменные посты едва выступают над снегом у древней дороги. Повреждённые часовые охраняют проходы, которыми пытаются воспользоваться искатели реликвий.',6,26),
 ('khar_delba','behemoth_crater','Кратер боевых зверей','Высокие стенки скрывают от ветра широкую промёрзшую впадину. Здесь встречаются огромные боевые звери, а у её краёв охотятся более мелкие хищники.',7,31),
 ('khar_delba','citadel_outer_approach','Подступы к древней цитадели','Между ледяными грядами проходит дорога к планетарному укреплению древних ситхов. Расхитители избегают крупных зверей и обходят уцелевшие механические посты.',8,36),
 ('khar_delba','frozen_citadel_courts','Промёрзшие дворы цитадели','Снег лежит между разрушенными стенами и основаниями башен. Среди широких дворов бродят боевые звери, а у закрытых проходов действуют старые часовые.',9,41),
 ('khar_delba','inner_citadel_ruins','Внутренние руины цитадели','Обвалившиеся галереи окружают уцелевшие каменные залы. Расхитители укрепились у найденных хранилищ, пока механические стражи и огромные звери удерживают остальные подходы.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('chandrila','rural_landing','chandrila_crop_raider','chandrila_harvest_droid','chandrila_crop_raider'),
 ('chandrila','forest_margins','chandrila_wildlife_smuggler','chandrila_crop_raider','chandrila_wildlife_smuggler'),
 ('chandrila','silver_sea_coves','chandrila_crop_raider','rogue_loader','chandrila_wildlife_smuggler'),
 ('chandrila','orchard_storehouses','chandrila_crop_raider','chandrila_harvest_droid','chandrila_harvest_droid'),
 ('chandrila','crystal_canyon_trails','chandrila_wildlife_smuggler','chandrila_estate_sentry','chandrila_wildlife_smuggler'),
 ('chandrila','irrigation_terraces','chandrila_harvest_droid','chandrila_crop_raider','chandrila_harvest_droid'),
 ('chandrila','illegal_capture_compound','chandrila_wildlife_smuggler','chandrila_estate_sentry','chandrila_wildlife_smuggler'),
 ('chandrila','old_country_estate','chandrila_crop_raider','chandrila_estate_sentry','chandrila_estate_sentry'),
 ('chandrila','closed_agricultural_station','chandrila_harvest_droid','chandrila_estate_sentry','chandrila_harvest_droid'),
 ('chandrila','hidden_manor','chandrila_crop_raider','chandrila_wildlife_smuggler','chandrila_estate_sentry'),
 ('duro','surface_freight_yard','duro_salvage_raider','rogue_loader','duro_reclamation_droid'),
 ('duro','restoration_blocks','duro_reclamation_droid','duro_salvage_raider','duro_reclamation_droid'),
 ('duro','scrap_sorting_fields','duro_salvage_raider','duro_reclamation_droid','duro_contraband_technician'),
 ('duro','old_power_corridors','duro_contraband_technician','duro_reclamation_droid','duro_industrial_warden'),
 ('duro','sealed_processing_halls','duro_reclamation_droid','duro_salvage_raider','duro_industrial_warden'),
 ('duro','reclaimed_factory_yards','duro_salvage_raider','duro_reclamation_droid','duro_contraband_technician'),
 ('duro','underground_workshops','duro_contraband_technician','duro_reclamation_droid','duro_industrial_warden'),
 ('duro','industrial_record_vault','duro_contraband_technician','duro_industrial_warden','duro_contraband_technician'),
 ('duro','automated_security_depot','duro_industrial_warden','duro_contraband_technician','duro_industrial_warden'),
 ('duro','seized_production_complex','duro_salvage_raider','duro_contraband_technician','duro_industrial_warden'),
 ('brentaal_iv','traders_landing','brentaal_freight_raider','brentaal_cargo_automaton','brentaal_customs_forger'),
 ('brentaal_iv','valley_warehouse_rows','brentaal_freight_raider','brentaal_cargo_automaton','brentaal_freight_raider'),
 ('brentaal_iv','coastal_freight_road','brentaal_freight_raider','brentaal_kundril','brentaal_kundril'),
 ('brentaal_iv','canyon_transit_yard','brentaal_freight_raider','brentaal_cargo_automaton','brentaal_customs_forger'),
 ('brentaal_iv','mountain_warehouse_steps','brentaal_freight_raider','brentaal_cargo_automaton','brentaal_customs_forger'),
 ('brentaal_iv','abandoned_sea_terminal','brentaal_kundril','brentaal_freight_raider','brentaal_kundril'),
 ('brentaal_iv','forged_manifest_archive','brentaal_customs_forger','brentaal_cargo_automaton','brentaal_customs_forger'),
 ('brentaal_iv','closed_plateau_port','brentaal_freight_raider','brentaal_cargo_automaton','brentaal_customs_forger'),
 ('brentaal_iv','seacliff_loading_galleries','brentaal_freight_raider','brentaal_cargo_automaton','brentaal_kundril'),
 ('brentaal_iv','smugglers_distribution_hub','brentaal_freight_raider','brentaal_customs_forger','brentaal_cargo_automaton'),
 ('jaguada','desert_supply_landing','jaguada_relic_smuggler','jaguada_dune_stalker','jaguada_excavation_droid'),
 ('jaguada','dune_tracks','jaguada_dune_stalker','jaguada_relic_smuggler','jaguada_dune_stalker'),
 ('jaguada','abandoned_wells','jaguada_dune_stalker','jaguada_relic_smuggler','jaguada_relic_smuggler'),
 ('jaguada','buried_outer_walls','jaguada_relic_smuggler','jaguada_excavation_droid','jaguada_relic_smuggler'),
 ('jaguada','dry_excavation_trenches','jaguada_relic_smuggler','jaguada_dune_stalker','jaguada_excavation_droid'),
 ('jaguada','caravan_hideout','jaguada_relic_smuggler','contraband_guard','jaguada_renegade_adept'),
 ('jaguada','sandstone_galleries','jaguada_relic_smuggler','jaguada_excavation_droid','jaguada_renegade_adept'),
 ('jaguada','sealed_relic_chambers','jaguada_renegade_adept','jaguada_excavation_droid','jaguada_renegade_adept'),
 ('jaguada','broken_sanctuary_steps','jaguada_renegade_adept','jaguada_excavation_droid','jaguada_renegade_adept'),
 ('jaguada','hidden_desert_enclave','jaguada_relic_smuggler','jaguada_renegade_adept','jaguada_renegade_adept'),
 ('rhelg','farmland_landing','rhelg_supply_raider','rhelg_field_stalker','rhelg_harvest_droid'),
 ('rhelg','overgrown_field_boundaries','rhelg_field_stalker','rhelg_supply_raider','rhelg_field_stalker'),
 ('rhelg','old_granaries','rhelg_supply_raider','rhelg_harvest_droid','rhelg_supply_raider'),
 ('rhelg','drainage_channels','rhelg_field_stalker','rhelg_harvest_droid','rhelg_harvest_droid'),
 ('rhelg','abandoned_farmworks','rhelg_supply_raider','rhelg_harvest_droid','rhelg_harvest_droid'),
 ('rhelg','estate_supply_road','rhelg_supply_raider','rhelg_palace_guardian','rhelg_supply_raider'),
 ('rhelg','overgrown_estate_courts','rhelg_field_stalker','rhelg_supply_raider','rhelg_palace_guardian'),
 ('rhelg','kressh_palace_approach','rhelg_supply_raider','rhelg_palace_guardian','rhelg_palace_guardian'),
 ('rhelg','palace_service_passages','rhelg_supply_raider','rhelg_palace_guardian','rhelg_supply_raider'),
 ('rhelg','sealed_palace_wing','rhelg_supply_raider','rhelg_palace_guardian','rhelg_palace_guardian'),
 ('khar_delba','sheltered_ice_landing','khar_delba_relic_raider','khar_delba_ice_stalker','khar_delba_relic_raider'),
 ('khar_delba','snow_hollows','khar_delba_ice_stalker','khar_delba_relic_raider','khar_delba_ice_stalker'),
 ('khar_delba','frozen_crater_edge','khar_delba_ice_stalker','khar_delba_relic_raider','khar_delba_ice_stalker'),
 ('khar_delba','abandoned_expedition_camp','khar_delba_relic_raider','khar_delba_sentry_droid','khar_delba_relic_raider'),
 ('khar_delba','icebound_caverns','cave_predator','khar_delba_ice_stalker','khar_delba_relic_raider'),
 ('khar_delba','buried_watchworks','khar_delba_sentry_droid','khar_delba_relic_raider','khar_delba_sentry_droid'),
 ('khar_delba','behemoth_crater','khar_delba_ice_stalker','khar_delba_war_behemoth','khar_delba_war_behemoth'),
 ('khar_delba','citadel_outer_approach','khar_delba_relic_raider','khar_delba_sentry_droid','khar_delba_war_behemoth'),
 ('khar_delba','frozen_citadel_courts','khar_delba_sentry_droid','khar_delba_relic_raider','khar_delba_war_behemoth'),
 ('khar_delba','inner_citadel_ruins','khar_delba_relic_raider','khar_delba_sentry_droid','khar_delba_war_behemoth')
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
WHERE p.code IN ('chandrila','duro','brentaal_iv','jaguada','rhelg','khar_delba') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
