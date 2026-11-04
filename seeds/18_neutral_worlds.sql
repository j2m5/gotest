-- Second neutral-world batch. Content does not change faction ownership or access.
-- Early Great Galactic War; districts and enemy roles are authored adaptations.
-- No later Makeb occupation, Clone Wars armies or Imperial/Vong devastation.
-- Kashyyyk and Trandosha do not include their moons. Ryloth avoids disputed rotation.
-- Sullust uses habitable tunnels; Ithor uses city decks above the protected jungle.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('makeb_makrin','Разъярённый макрин','Обитает на травянистых уступах и в пещерах Макеба. Защищает привычные укрытия от существ, приближающихся к нему среди скал.','beast'),
 ('makeb_cliff_stalker','Хищник макебских ущелий','Держится у каменных троп и заросших склонов высоких плато. Подстерегает добычу на узких переходах между пещерами.','beast'),
 ('makeb_ore_raider','Похититель макебской руды','Грабит небольшие добывающие площадки и перехватывает минеральные грузы. Охраняет краденое в закрытых шахтных дворах.','humanoid'),
 ('makeb_mining_droid','Повреждённый горный дроид','Остался на заброшенном участке добычи Макеба. Ошибки распознавания заставляют его применять буровые инструменты против посетителей.','droid'),
 ('nal_hutta_gang_enforcer','Вышибала болотной банды','Вооружённый участник группы, захватившей старые складские дворы. Отнимает припасы у перевозчиков и защищает тайники своей банды.','humanoid'),
 ('nal_hutta_chemilizard','Химическая ящерица','Приспособилась к загрязнённым топям и промышленным стокам Нал-Хатты. Подбирается к добыче через вязкую грязь и заросли болотной травы.','beast'),
 ('nal_hutta_refinery_droid','Сбойный перерабатывающий дроид','Обслуживал насосы и линии старого перерабатывающего комплекса. Повреждённая программа заставляет его устранять движущиеся помехи рабочими захватами.','droid'),
 ('nal_hutta_contraband_guard','Охранник контрабандных складов','Нанят преступной сетью для защиты нелегальных грузов. Контролирует проходы к тайным пристаням и укреплённым хранилищам.','humanoid'),
 ('kashyyyk_forest_kinrath','Лесной кинрат','Скрывается среди корней и густой растительности Кашиика. Нападает на существ, оказавшихся возле его укрытия или гнезда.','beast'),
 ('kashyyyk_katarn','Катарн','Хищник нижних лесных ярусов Кашиика охотится в постоянной тени. Использует переплетения корней и поваленные стволы для внезапного нападения.','beast'),
 ('kashyyyk_webweaver','Кашиикский паук-плетельщик','Обитает в тёмных участках леса и каменных полостях под корнями. Выслеживает добычу возле своих паутинных укрытий.','beast'),
 ('kashyyyk_slaver','Работорговец лесных окраин','Участник преступной группы, похищающей путников вдали от поселений. Охраняет скрытые лагеря и площадки для вывоза пленников.','humanoid'),
 ('trandosha_thicket_stalker','Хищник трандошанских зарослей','Держится на границе влажных лесов и открытых равнин. Выслеживает добычу у звериных троп, оврагов и заброшенных стоянок.','beast'),
 ('trandosha_captive_trader','Торговец пленниками','Содержит нелегальные перевалочные пункты в удалённых районах Трандоши. Вооружённо защищает свою преступную сеть и её транспортные запасы.','humanoid'),
 ('trandosha_outlaw_tracker','Следопыт преступной банды','Разыскивает беглецов и свидетелей для работорговцев. Контролирует скрытые тропы и устраивает засады на подходах к лагерям.','humanoid'),
 ('trandosha_trap_droid','Взломанный ловчий дроид','Преступники переделали рабочую машину для охраны своих стоянок. Дроид преследует посторонних и применяет силовые захваты.','droid'),
 ('ryloth_gutkurr','Гаткур','Быстрый хищник бродит по каменистым долинам Рилота. Подбирается к добыче вдоль скал и защищает укрытия под природными навесами.','beast'),
 ('ryloth_lylek','Лайлек','Крупный панцирный хищник занимает удалённые ущелья и пещеры Рилота. Его мощные конечности и щупальца делают опасным даже краткое сближение.','beast'),
 ('ryloth_supply_raider','Налётчик пещерных дорог','Грабит караваны снабжения между подземными поселениями. Прячет захваченные грузы в покинутых складских нишах и каменных дворах.','humanoid'),
 ('ryloth_tunnel_droid','Неисправный тоннельный дроид','Обслуживал выработки и грузовые проходы в скалах Рилота. После сбоя атакует посетителей рабочими инструментами и подъёмными захватами.','droid'),
 ('rodia_ghest','Гест','Крупный родианский хищник скрывается в изолированных болотных чащах. Охраняет своё логово и подстерегает добычу у вязких берегов.','beast'),
 ('rodia_poacher','Браконьер родианских болот','Расставляет ловушки вдали от поселений и охраняемых природных участков. Защищает укрытия с незаконно добытыми животными и снаряжением.','humanoid'),
 ('rodia_cargo_raider','Налётчик родианских причалов','Перехватывает грузы на удалённых водных торговых путях. Использует пустующие причалы и мастерские для хранения добычи.','humanoid'),
 ('rodia_service_droid','Повреждённый причальный дроид','Ремонтировал настилы и перевозил оборудование между складами. Сбой наведения заставляет его применять инструменты против движущихся целей.','droid'),
 ('sullust_cave_stalker','Пещерный хищник Салласта','Обитает в прохладных полостях под вулканической поверхностью. Подстерегает добычу у подземных водоёмов и тёмных ответвлений тоннелей.','beast'),
 ('sullust_ore_thief','Похититель подземных грузов','Грабит перевозчиков сырья и оборудования в удалённых районах Салласта. Охраняет награбленное в закрытых производственных корпусах.','humanoid'),
 ('sullust_mining_droid','Сбойный салластский проходчик','Остался в старых выработках после остановки добычи. Неисправные датчики направляют тяжёлые рабочие инструменты на живых посетителей.','droid'),
 ('sullust_security_droid','Захваченный промышленный страж','Перепрограммирован похитителями производственного оборудования. Контролирует складские галереи и не пропускает посторонних к преступным тайникам.','droid'),
 ('ithor_cargo_raider','Налётчик грузовых палуб','Проник в закрытые служебные сектора парящего города Итора. Использует пустующие причалы для хранения и вывоза украденных грузов.','humanoid'),
 ('ithor_plant_smuggler','Контрабандист ботанических образцов','Похищает семена и саженцы из городских коллекций. Охраняет незаконные склады и оборудование, подготовленное для тайной перевозки.','humanoid'),
 ('ithor_service_droid','Повреждённый садовый техдроид','Обслуживал городские питомники и сети подачи воды. После сбоя применяет ремонтные инструменты против посетителей закрытых помещений.','droid'),
 ('ithor_security_droid','Взломанный страж хранилища','Преступники приспособили охранную машину для защиты похищенных образцов. Патрулирует складские палубы и атакует посторонних.','droid')
ON CONFLICT (code) DO NOTHING;

-- Makeb and Nal Hutta: mineral plateaus and polluted industrial wetlands.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('makeb','plateau_landing','Посадочная терраса плато','Широкий каменный уступ служит площадкой для небольших грузовых судов. У заброшенных навесов скрываются похитители руды и оставленные горные машины.',1,1),
 ('makeb','grassland_rim_path','Тропа травянистого края','Между скальными выступами тянется полоса густой травы. Макрины занимают защищённые ложбины, а местные хищники следят за узкими переходами.',2,6),
 ('makeb','mineral_caravan_road','Дорога минеральных караванов','Укреплённый путь соединяет старые площадки добычи на одном плато. Налётчики поджидают перевозчиков возле пустующих разгрузочных дворов.',3,11),
 ('makeb','abandoned_quarry','Покинутый каменный карьер','Ступени выработки спускаются под нависающие слои породы. Между складами инструмента бродят сбойные проходчики и охранники украденных грузов.',4,16),
 ('makeb','makrin_cave_shelters','Пещерные укрытия макринов','Заросшие входы ведут в прохладные залы внутри каменного столба. Макрины защищают привычные укрытия, а хищники держатся у боковых расщелин.',5,21),
 ('makeb','ore_sorting_yards','Дворы сортировки руды','Старые конвейеры окружают площадки с брошенными контейнерами. Преступная группа восстановила часть техники и устроила здесь склад краденого сырья.',6,26),
 ('makeb','high_cliff_passages','Проходы высоких скал','Тропа уходит под природные каменные арки вдоль глубокого ущелья. В зарослях и тенистых нишах скрываются опасные обитатели плато.',7,31),
 ('makeb','deep_mineral_galleries','Глубокие минеральные галереи','Под старым карьером сохранились длинные сухие выработки. Повреждённые горные машины продолжают работать рядом с тайниками похитителей руды.',8,36),
 ('makeb','seized_export_warehouse','Захваченный экспортный склад','Укреплённый корпус стоит у закрытой грузовой террасы. Налётчики собрали здесь дорогие минеральные партии и запустили рабочие автоматы для охраны.',9,41),
 ('makeb','hidden_mesa_depot','Тайная база среди каменных столбов','Покинутый добывающий комплекс скрыт за высоким краем плато. Вооружённая сеть похитителей контролирует мастерские, склады и единственный удобный подход.',10,46),
 ('nal_hutta','marshside_cargo_landing','Грузовая площадка у топей','Настил над грязью связывает небольшой причал со старыми складами. Местная банда заняла пустующие помещения и перехватывает проходящие грузы.',1,1),
 ('nal_hutta','stagnant_causeways','Дамбы застойных болот','Узкие земляные насыпи проходят между тёмными лужами и редкой травой. Химические ящерицы скрываются у кромки воды, а бандиты следят за тропами.',2,6),
 ('nal_hutta','old_scrap_yards','Дворы старых разборщиков','Кучи деталей окружают закрытые мастерские и ржавые подъёмники. Преступники сортируют похищенное оборудование рядом с неисправными перерабатывающими машинами.',3,11),
 ('nal_hutta','waste_channel_walkways','Настилы сточных каналов','Служебные мостки тянутся вдоль загрязнённых промышленных стоков. Среди труб появляются болотные хищники и оставленные без присмотра дроиды.',4,16),
 ('nal_hutta','contraband_pier','Контрабандная пристань','Покосившиеся склады скрыты за широким поясом топей. Вооружённая охрана контролирует мостки, по которым перевозят нелегальные грузы.',5,21),
 ('nal_hutta','idle_refinery_halls','Залы остановленного переработчика','Пустующие цеха соединены сухими техническими переходами. Часть машин всё ещё движется среди ёмкостей, присвоенных местной преступной сетью.',6,26),
 ('nal_hutta','chemilizard_backwaters','Затоны химических ящериц','Небольшие островки окружены вязкими протоками и больной болотной растительностью. В этих укрытых от дорог местах держатся крупные местные хищники.',7,31),
 ('nal_hutta','sealed_smuggler_stores','Закрытые хранилища контрабандистов','За укреплёнными дверями старого торгового корпуса спрятаны ценные партии товаров. Наёмная охрана и сбойная техника перекрывают служебные проходы.',8,36),
 ('nal_hutta','gang_service_compound','Служебный комплекс преступной сети','Огороженный двор объединяет ремонтные боксы и оружейные склады. Банда использует его для подготовки налётов и защиты основных запасов.',9,41),
 ('nal_hutta','fortified_marsh_exchange','Укреплённый узел болотной торговли','На прочном островном основании стоит скрытая перевалочная база. Её охрана контролирует все настилы к складам, где сходятся незаконные поставки.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Kashyyyk and Trandosha: forest strata, savannas and isolated criminal camps.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('kashyyyk','outer_canopy_landing','Внешний настил в кронах','Посадочная площадка расположена у старого грузового пути между деревьями. Вдали от жилых мостов появились чужие вооружённые дозоры и тайники похитителей.',1,1),
 ('kashyyyk','overgrown_branch_walkways','Заросшие мосты ветвей','Прочные ветви и деревянные переходы соединяют заброшенные хозяйственные настилы. Среди зелени прячутся лесные кинраты, а работорговцы следят за дальними подходами.',2,6),
 ('kashyyyk','old_lift_station','Старая подъёмная станция','Грузовые площадки спускаются к более тёмному ярусу леса. Преступники используют закрытые навесы для хранения снаряжения и подготовки похищений.',3,11),
 ('kashyyyk','lower_bough_paths','Тропы нижних ветвей','Солнечный свет едва пробивается между огромными стволами. Кинраты занимают растительные укрытия, а катарны выходят к широким переплетениям корней.',4,16),
 ('kashyyyk','hidden_captive_camp','Скрытый лагерь работорговцев','Несколько укреплённых настилов спрятаны за стеной лиан. Вооружённая группа охраняет незаконные клетки и путь к грузовому подъёмнику.',5,21),
 ('kashyyyk','root_arches','Арки гигантских корней','Под сводами корней начинаются тёмные наземные тропы. Катарны охотятся среди поваленных стволов, а кинраты скрываются в заросших полостях.',6,26),
 ('kashyyyk','webbed_forest_hollows','Паутинные лесные полости','Затенённые проходы покрыты плотными нитями между корнями и скалами. Пауки-плетельщики занимают укрытия и подстерегают добычу у входов.',7,31),
 ('kashyyyk','shadowland_riverbank','Берег реки Земель Теней','Тёмная река огибает густые скопления корней. Сухие береговые тропы проходят через охотничьи угодья катарнов и лесных членистоногих.',8,36),
 ('kashyyyk','deep_webweaver_caves','Глубокие пещеры плетельщиков','Каменные полости под корневым поясом почти полностью скрыты паутиной. Крупные пауки занимают внутренние залы и охраняют узкие выходы.',9,41),
 ('kashyyyk','ancient_root_depths','Глубины древнего леса','Массивные корни образуют лабиринт в постоянной тени. Самые опасные хищники окрестностей делят между собой укрытия и выходы к воде.',10,46),
 ('trandosha','woodland_trade_landing','Торговая площадка у леса','За грузовыми навесами начинается дорога к удалённым поселениям. В покинутых служебных дворах скрываются участники преступной сети ловцов.',1,1),
 ('trandosha','humid_thicket_trail','Тропа влажных зарослей','Плотная растительность скрывает края старой дороги. Местные хищники и вооружённые следопыты используют густую листву для засад.',2,6),
 ('trandosha','savanna_caravan_way','Караванный путь саванны','Открытая равнина пересечена редкими рощами и каменными ложбинами. Банда ловцов следит за путниками с пустующих стоянок.',3,11),
 ('trandosha','abandoned_trapper_yards','Покинутые дворы ловцов','Старые мастерские окружены клетками и разобранными подъёмниками. Преступники восстанавливают оборудование, оставляя охрану взломанным рабочим машинам.',4,16),
 ('trandosha','rocky_waterholes','Водопои среди скал','Цепь неглубоких водоёмов лежит на границе леса и сухой равнины. Хищники защищают подходы к воде, а следопыты прячутся на каменных уступах.',5,21),
 ('trandosha','illicit_holding_camp','Нелегальный лагерь содержания','За ограждением скрываются помещения преступной перевалочной базы. Торговцы пленниками и ловчие дроиды контролируют выходы к дороге.',6,26),
 ('trandosha','highland_tracking_paths','Следовые тропы нагорья','Каменистые проходы обходят обрывистые склоны и труднодоступные рощи. Здесь банда оставила наблюдателей, выслеживающих беглецов из своих лагерей.',7,31),
 ('trandosha','seized_equipment_store','Захваченный склад снаряжения','Пустующий хозяйственный корпус заполнен оружием, ловушками и крадеными деталями. Преступная охрана перекрыла подъезды и запустила патрульных автоматов.',8,36),
 ('trandosha','hidden_export_compound','Тайный перевалочный комплекс','Закрытый двор соединён с отдалённой грузовой площадкой. Работорговцы готовят незаконные перевозки под охраной опытных следопытов.',9,41),
 ('trandosha','outlaw_ridge_fort','Укрепление преступников на хребте','Старая каменная база скрыта за узким горным проходом. Руководители сети ловцов собрали здесь вооружённую охрану и переделанные машины.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Ryloth and Rodia: sheltered rocky routes, cave settlements and humid wetlands.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('ryloth','sheltered_cargo_landing','Укрытая грузовая площадка','Каменный навес защищает посадочный настил от сильных ветров. Старые складские ниши у входа заняты налётчиками и неисправной грузовой техникой.',1,1),
 ('ryloth','shadowed_valley_road','Дорога тенистой долины','Тропа проходит вдоль высоких скал, закрывающих открытые пустоши. Гаткуры выходят из боковых расщелин, а грабители следят за караванами снабжения.',2,6),
 ('ryloth','outer_cavern_stores','Склады у входа в пещеры','Покинутые грузовые помещения отделены от жилых районов длинным проходом. Налётчики используют старые кладовые для хранения захваченных припасов.',3,11),
 ('ryloth','old_turbine_galleries','Галереи старых турбин','Служебные коридоры проходят внутри скального массива рядом с ветровыми установками. Сбойные машины перекрывают переходы, где преступники спрятали своё оборудование.',4,16),
 ('ryloth','gutkurr_ravines','Овраги гаткуров','За скальными арками открывается цепь защищённых от ветра низин. Здесь гаткуры занимают природные укрытия и охраняют подходы к ним.',5,21),
 ('ryloth','abandoned_supply_station','Покинутая станция снабжения','Закрытый каменный комплекс когда-то обслуживал дальние поселения. Теперь его дворы заняты грабителями, присвоившими склады и рабочие автоматы.',6,26),
 ('ryloth','deep_rock_shelters','Глубокие скальные убежища','Тесные проходы уходят в просторные пещеры вдали от жилых районов. Гаткуры и крупные лайлеки используют эти укрытия как охотничьи логовища.',7,31),
 ('ryloth','buried_cargo_tunnels','Заглублённые грузовые тоннели','Старые перевозочные пути соединяют закрытые складские корпуса. Налётчики укрепили развилки и используют повреждённые машины для защиты добычи.',8,36),
 ('ryloth','hidden_raider_fort','Тайное укрепление налётчиков','За каменными воротами расположен склад оружия и захваченных товаров. Преступная группа контролирует все подходы через окружающие пещеры.',9,41),
 ('ryloth','lylek_cavern_depths','Глубинные логовища лайлеков','Массивные каменные своды окружают удалённую сеть природных залов. Крупные панцирные хищники защищают свои логовища и узкие проходы между ними.',10,46),
 ('rodia','riverport_outskirts','Окраины речного порта','Грузовые настилы заканчиваются у заболоченных городских окраин. В закрытых мастерских прячутся налётчики и оставленные без обслуживания машины.',1,1),
 ('rodia','canalside_warehouses','Склады вдоль канала','Старые погрузочные дворы тянутся вдоль водного торгового пути. Преступники перегружают похищенные товары за стенами пустующих корпусов.',2,6),
 ('rodia','jungle_causeway','Дамба среди джунглей','Укреплённая дорога пересекает влажные заросли и неглубокие протоки. Браконьеры используют заброшенные стоянки, а налётчики наблюдают за транспортом.',3,11),
 ('rodia','idle_waterfront_workshops','Пустующие прибрежные мастерские','Оборудованные дворы отделены от действующих причалов заросшими стенами. Сбойные техдроиды продолжают работу рядом с тайниками похитителей грузов.',4,16),
 ('rodia','poacher_reed_camp','Браконьерская стоянка в тростниках','За густой растительностью скрыты клетки, лодочные настилы и склад ловушек. Вооружённая группа охраняет незаконную добычу и пути вывоза.',5,21),
 ('rodia','seized_trading_annex','Захваченный торговый корпус','Закрытое крыло старого комплекса стало базой речных грабителей. Между складами и ремонтными помещениями выставлены дозоры.',6,26),
 ('rodia','isolated_marsh_paths','Тропы отдалённых болот','Узкие сухие полосы проходят между густыми зарослями и тёмными заводями. В этой удалённой части леса сохранились логовища крупных гестов.',7,31),
 ('rodia','hidden_trapper_enclosures','Скрытые загоны ловцов','Укреплённые настилы окружают незаконные клетки и склады снаряжения. Браконьеры защищают перевалочную базу вместе с нанятыми речными налётчиками.',8,36),
 ('rodia','deep_swamp_hollows','Глубокие болотные чащи','Старые деревья поднимаются над изолированными участками топи. Гесты занимают укрытия среди корней и не подпускают чужаков к своим логовищам.',9,41),
 ('rodia','fortified_river_hideout','Укреплённое речное убежище','Сеть преступных перевозчиков заняла покинутый портовый комплекс. Здесь хранят основные грузы, ремонтируют технику и готовят новые налёты.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Sullust's habitable underground; Ithor's unused city sectors above Mother Jungle.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('sullust','cavern_port_landing','Посадочный зал пещерного порта','Закрытая грузовая площадка расположена в обитаемом подземном поясе. В пустующих служебных помещениях скрываются похитители оборудования и сбойные машины.',1,1),
 ('sullust','old_freight_galleries','Старые грузовые галереи','Широкие тоннели соединяют заброшенные погрузочные дворы. Грабители используют боковые склады, а рабочие автоматы продолжают обходить тёмные участки.',2,6),
 ('sullust','cool_cavern_paths','Тропы прохладных пещер','Природные проходы лежат вдали от горячих трещин верхней коры. У подземных ручьёв и зарослей лишайника скрываются местные хищники.',3,11),
 ('sullust','abandoned_processing_halls','Заброшенные обогатительные залы','Сухие производственные помещения отделены от действующих предприятий каменными перегородками. Похитители металла заняли склады рядом с неисправными проходчиками.',4,16),
 ('sullust','underground_lake_margin','Берег подземного озера','Каменистая тропа огибает прохладный водоём внутри обширной пещеры. Хищники занимают боковые полости, а грабители прячут припасы у старых мостков.',5,21),
 ('sullust','closed_tool_foundry','Закрытая инструментальная мастерская','Подземный корпус хранит остатки производственных линий и тяжёлых инструментов. Преступники восстановили часть техники и перепрошили охранных автоматов.',6,26),
 ('sullust','deep_survey_tunnels','Глубокие разведочные тоннели','Старые выработки выходят в естественные полости с пригодным для дыхания воздухом. Между брошенными приборами движутся пещерные звери и повреждённые машины.',7,31),
 ('sullust','seized_machine_depot','Захваченный склад машин','Закрытые ремонтные галереи превращены в хранилище украденного оборудования. Охранные автоматы перекрывают проходы к площадкам сборки и погрузки.',8,36),
 ('sullust','hidden_industrial_vaults','Тайные промышленные хранилища','За стенами старого административного корпуса устроены укреплённые склады. Преступная сеть держит здесь ценные детали и документы нелегальных поставок.',9,41),
 ('sullust','outlaw_subterranean_complex','Подземный комплекс налётчиков','Изолированные производственные залы соединены длинными закрытыми переходами. Грабители собрали здесь основные запасы и самые опасные охранные машины.',10,46),
 ('ithor','outer_city_cargo_deck','Внешняя грузовая палуба','Парящий город проходит над зелёными лесами Итора, не касаясь их поверхности. На закрытых грузовых настилах его старого сектора обосновались похитители товаров.',1,1),
 ('ithor','unused_supply_walkways','Неиспользуемые служебные настилы','Переходы вдоль внешних ограждений связывают пустующие складские помещения. Налётчики следят за подходами, а повреждённые машины ремонтируют давно закрытые участки.',2,6),
 ('ithor','old_nursery_annex','Старое крыло питомников','Часть городского ботанического комплекса выведена из эксплуатации. Похитители образцов устроили тайники среди пустых стеллажей и садового оборудования.',3,11),
 ('ithor','water_service_galleries','Галереи подачи воды','Сухие технические проходы идут вдоль резервуаров и распределительных труб. Сбойные садовые дроиды перекрывают пути к закрытым служебным помещениям.',4,16),
 ('ithor','vacant_conservatory','Пустующий оранжерейный корпус','Старые посадочные залы отделены от действующих городских садов. Контрабандисты заняли складскую часть и используют её для упаковки украденных растений.',5,21),
 ('ithor','seized_seed_archive','Захваченный архив семян','В закрытом исследовательском крыле сохранились защищённые комнаты хранения. Преступная группа прячет здесь похищенные образцы под охраной перепрошитых машин.',6,26),
 ('ithor','lower_filter_deck','Нижняя палуба фильтров','Обслуживаемые изнутри установки расположены в сухих секциях городского корпуса. Повреждённые автоматы и преступные дозоры контролируют тесные переходы.',7,31),
 ('ithor','illicit_packing_workshops','Подпольные упаковочные мастерские','Пустующие ремонтные помещения превращены в центр подготовки незаконных грузов. Налётчики собирают контейнеры, а контрабандисты сортируют похищенные образцы.',8,36),
 ('ithor','hidden_botanical_stores','Тайные ботанические хранилища','Закрытые складские комнаты соединены защищёнными служебными коридорами. Контрабандная сеть сосредоточила здесь ценные коллекции и усиленную охрану.',9,41),
 ('ithor','smuggler_export_hangar','Тайный ангар вывоза образцов','Заброшенный грузовой ангар расположен в конце нежилого сектора парящего города. Преступники укрепили подходы и готовят похищенные грузы к отправке за пределы Итора.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('makeb','plateau_landing','makeb_ore_raider','makeb_mining_droid','makeb_ore_raider'),
 ('makeb','grassland_rim_path','makeb_makrin','makeb_cliff_stalker','makeb_makrin'),
 ('makeb','mineral_caravan_road','makeb_ore_raider','makeb_mining_droid','makeb_ore_raider'),
 ('makeb','abandoned_quarry','makeb_mining_droid','makeb_ore_raider','makeb_mining_droid'),
 ('makeb','makrin_cave_shelters','makeb_makrin','makeb_cliff_stalker','makeb_makrin'),
 ('makeb','ore_sorting_yards','makeb_ore_raider','makeb_mining_droid','makeb_ore_raider'),
 ('makeb','high_cliff_passages','makeb_cliff_stalker','makeb_makrin','makeb_cliff_stalker'),
 ('makeb','deep_mineral_galleries','makeb_mining_droid','makeb_ore_raider','makeb_mining_droid'),
 ('makeb','seized_export_warehouse','makeb_ore_raider','makeb_mining_droid','makeb_ore_raider'),
 ('makeb','hidden_mesa_depot','makeb_ore_raider','makeb_mining_droid','makeb_ore_raider'),
 ('nal_hutta','marshside_cargo_landing','nal_hutta_gang_enforcer','rogue_loader','nal_hutta_gang_enforcer'),
 ('nal_hutta','stagnant_causeways','nal_hutta_gang_enforcer','nal_hutta_chemilizard','nal_hutta_chemilizard'),
 ('nal_hutta','old_scrap_yards','nal_hutta_gang_enforcer','nal_hutta_refinery_droid','nal_hutta_refinery_droid'),
 ('nal_hutta','waste_channel_walkways','nal_hutta_refinery_droid','nal_hutta_chemilizard','nal_hutta_chemilizard'),
 ('nal_hutta','contraband_pier','nal_hutta_gang_enforcer','nal_hutta_contraband_guard','nal_hutta_contraband_guard'),
 ('nal_hutta','idle_refinery_halls','nal_hutta_refinery_droid','nal_hutta_gang_enforcer','nal_hutta_refinery_droid'),
 ('nal_hutta','chemilizard_backwaters','nal_hutta_chemilizard','nal_hutta_chemilizard','nal_hutta_chemilizard'),
 ('nal_hutta','sealed_smuggler_stores','nal_hutta_contraband_guard','nal_hutta_refinery_droid','nal_hutta_contraband_guard'),
 ('nal_hutta','gang_service_compound','nal_hutta_gang_enforcer','nal_hutta_refinery_droid','nal_hutta_contraband_guard'),
 ('nal_hutta','fortified_marsh_exchange','nal_hutta_gang_enforcer','nal_hutta_contraband_guard','nal_hutta_contraband_guard'),
 ('kashyyyk','outer_canopy_landing','kashyyyk_slaver','kashyyyk_slaver','kashyyyk_slaver'),
 ('kashyyyk','overgrown_branch_walkways','kashyyyk_forest_kinrath','kashyyyk_slaver','kashyyyk_forest_kinrath'),
 ('kashyyyk','old_lift_station','kashyyyk_slaver','kashyyyk_slaver','kashyyyk_slaver'),
 ('kashyyyk','lower_bough_paths','kashyyyk_forest_kinrath','kashyyyk_katarn','kashyyyk_katarn'),
 ('kashyyyk','hidden_captive_camp','kashyyyk_slaver','kashyyyk_slaver','kashyyyk_slaver'),
 ('kashyyyk','root_arches','kashyyyk_forest_kinrath','kashyyyk_katarn','kashyyyk_katarn'),
 ('kashyyyk','webbed_forest_hollows','kashyyyk_webweaver','kashyyyk_webweaver','kashyyyk_webweaver'),
 ('kashyyyk','shadowland_riverbank','kashyyyk_forest_kinrath','kashyyyk_katarn','kashyyyk_webweaver'),
 ('kashyyyk','deep_webweaver_caves','kashyyyk_webweaver','kashyyyk_webweaver','kashyyyk_webweaver'),
 ('kashyyyk','ancient_root_depths','kashyyyk_forest_kinrath','kashyyyk_katarn','kashyyyk_webweaver'),
 ('trandosha','woodland_trade_landing','trandosha_outlaw_tracker','trandosha_trap_droid','trandosha_captive_trader'),
 ('trandosha','humid_thicket_trail','trandosha_thicket_stalker','trandosha_outlaw_tracker','trandosha_thicket_stalker'),
 ('trandosha','savanna_caravan_way','trandosha_outlaw_tracker','trandosha_thicket_stalker','trandosha_captive_trader'),
 ('trandosha','abandoned_trapper_yards','trandosha_outlaw_tracker','trandosha_trap_droid','trandosha_captive_trader'),
 ('trandosha','rocky_waterholes','trandosha_thicket_stalker','trandosha_outlaw_tracker','trandosha_thicket_stalker'),
 ('trandosha','illicit_holding_camp','trandosha_captive_trader','trandosha_trap_droid','trandosha_captive_trader'),
 ('trandosha','highland_tracking_paths','trandosha_outlaw_tracker','trandosha_thicket_stalker','trandosha_outlaw_tracker'),
 ('trandosha','seized_equipment_store','trandosha_captive_trader','trandosha_trap_droid','trandosha_trap_droid'),
 ('trandosha','hidden_export_compound','trandosha_outlaw_tracker','trandosha_captive_trader','trandosha_captive_trader'),
 ('trandosha','outlaw_ridge_fort','trandosha_outlaw_tracker','trandosha_captive_trader','trandosha_trap_droid'),
 ('ryloth','sheltered_cargo_landing','ryloth_supply_raider','ryloth_tunnel_droid','ryloth_supply_raider'),
 ('ryloth','shadowed_valley_road','ryloth_gutkurr','ryloth_supply_raider','ryloth_gutkurr'),
 ('ryloth','outer_cavern_stores','ryloth_supply_raider','ryloth_tunnel_droid','ryloth_supply_raider'),
 ('ryloth','old_turbine_galleries','ryloth_tunnel_droid','ryloth_supply_raider','ryloth_tunnel_droid'),
 ('ryloth','gutkurr_ravines','ryloth_gutkurr','ryloth_gutkurr','ryloth_gutkurr'),
 ('ryloth','abandoned_supply_station','ryloth_supply_raider','ryloth_tunnel_droid','ryloth_supply_raider'),
 ('ryloth','deep_rock_shelters','ryloth_gutkurr','ryloth_lylek','ryloth_lylek'),
 ('ryloth','buried_cargo_tunnels','ryloth_supply_raider','ryloth_tunnel_droid','ryloth_tunnel_droid'),
 ('ryloth','hidden_raider_fort','ryloth_supply_raider','ryloth_tunnel_droid','ryloth_supply_raider'),
 ('ryloth','lylek_cavern_depths','ryloth_gutkurr','ryloth_lylek','ryloth_lylek'),
 ('rodia','riverport_outskirts','rodia_cargo_raider','rodia_service_droid','rodia_cargo_raider'),
 ('rodia','canalside_warehouses','rodia_cargo_raider','rogue_loader','rodia_cargo_raider'),
 ('rodia','jungle_causeway','rodia_poacher','rodia_cargo_raider','rodia_poacher'),
 ('rodia','idle_waterfront_workshops','rodia_service_droid','rodia_cargo_raider','rodia_service_droid'),
 ('rodia','poacher_reed_camp','rodia_poacher','rodia_cargo_raider','rodia_poacher'),
 ('rodia','seized_trading_annex','rodia_cargo_raider','rodia_service_droid','rodia_cargo_raider'),
 ('rodia','isolated_marsh_paths','rodia_ghest','rodia_ghest','rodia_ghest'),
 ('rodia','hidden_trapper_enclosures','rodia_poacher','rodia_cargo_raider','rodia_poacher'),
 ('rodia','deep_swamp_hollows','rodia_ghest','rodia_ghest','rodia_ghest'),
 ('rodia','fortified_river_hideout','rodia_cargo_raider','rodia_service_droid','rodia_cargo_raider'),
 ('sullust','cavern_port_landing','sullust_ore_thief','sullust_mining_droid','sullust_ore_thief'),
 ('sullust','old_freight_galleries','sullust_ore_thief','sullust_mining_droid','sullust_ore_thief'),
 ('sullust','cool_cavern_paths','sullust_cave_stalker','sullust_cave_stalker','sullust_cave_stalker'),
 ('sullust','abandoned_processing_halls','sullust_ore_thief','sullust_mining_droid','sullust_mining_droid'),
 ('sullust','underground_lake_margin','sullust_cave_stalker','sullust_ore_thief','sullust_cave_stalker'),
 ('sullust','closed_tool_foundry','sullust_ore_thief','sullust_mining_droid','sullust_security_droid'),
 ('sullust','deep_survey_tunnels','sullust_cave_stalker','sullust_mining_droid','sullust_cave_stalker'),
 ('sullust','seized_machine_depot','sullust_ore_thief','sullust_security_droid','sullust_security_droid'),
 ('sullust','hidden_industrial_vaults','sullust_ore_thief','sullust_security_droid','sullust_security_droid'),
 ('sullust','outlaw_subterranean_complex','sullust_ore_thief','sullust_mining_droid','sullust_security_droid'),
 ('ithor','outer_city_cargo_deck','ithor_cargo_raider','rogue_loader','ithor_cargo_raider'),
 ('ithor','unused_supply_walkways','ithor_cargo_raider','ithor_service_droid','ithor_cargo_raider'),
 ('ithor','old_nursery_annex','ithor_plant_smuggler','ithor_service_droid','ithor_plant_smuggler'),
 ('ithor','water_service_galleries','ithor_service_droid','ithor_service_droid','ithor_service_droid'),
 ('ithor','vacant_conservatory','ithor_plant_smuggler','ithor_service_droid','ithor_plant_smuggler'),
 ('ithor','seized_seed_archive','ithor_plant_smuggler','ithor_security_droid','ithor_security_droid'),
 ('ithor','lower_filter_deck','ithor_service_droid','ithor_cargo_raider','ithor_security_droid'),
 ('ithor','illicit_packing_workshops','ithor_cargo_raider','ithor_plant_smuggler','ithor_security_droid'),
 ('ithor','hidden_botanical_stores','ithor_plant_smuggler','ithor_security_droid','ithor_security_droid'),
 ('ithor','smuggler_export_hangar','ithor_cargo_raider','ithor_plant_smuggler','ithor_security_droid')
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
WHERE p.code IN ('makeb','nal_hutta','kashyyyk','trandosha','ryloth','rodia','sullust','ithor') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
