package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/config"
	"github.com/ailms/backend/internal/infrastructure/database"
	infraRepos "github.com/ailms/backend/internal/infrastructure/repositories"
	"github.com/ailms/backend/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ── template types ─────────────────────────────────────────────────────────────

type qDef struct {
	q, correct string
	wrong      [3]string
}

type lessonDef struct {
	title, content, video, file, link string
	typ                               entities.LessonType
	quiz                              []qDef
}

type modDef struct {
	title   string
	lessons [4]lessonDef
}

type aDef struct {
	title, desc string
	pts, days   int
}

type eDef struct {
	title, desc       string
	mins, maxAttempts int
	pts               int
	qs                []qDef
}

type annDef struct {
	title, body string
}

type courseDef struct {
	title, desc string
	status      entities.CourseStatus
	mods        [3]modDef
	assigns     [2]aDef
	exams       [2]eDef
	anns        []annDef
}

type uDef struct {
	name, email string
	role        entities.Role
}

// ── seedDeps ──────────────────────────────────────────────────────────────────

type seedDeps struct {
	userRepo     *infraRepos.GORMUserRepository
	orgRepo      *infraRepos.GORMOrganizationRepository
	memberRepo   *infraRepos.GORMMembershipRepository
	courseRepo   *infraRepos.GORMCourseRepository
	ctRepo       *infraRepos.GORMCourseTeacherRepository
	moduleRepo   *infraRepos.GORMModuleRepository
	lessonRepo   *infraRepos.GORMLessonRepository
	quizRepo     *infraRepos.GORMQuizRepository
	qaRepo       *infraRepos.GORMQuizAttemptRepository
	assignRepo   *infraRepos.GORMAssignmentRepository
	examRepo     *infraRepos.GORMExamRepository
	eaRepo       *infraRepos.GORMExamAttemptRepository
	grantRepo    *infraRepos.GORMExtraAttemptGrantRepository
	groupRepo    *infraRepos.GORMGroupRepository
	announceRepo *infraRepos.GORMAnnouncementRepository
	progressRepo *infraRepos.GORMProgressRepository
}

// ── course templates ───────────────────────────────────────────────────────────

var courseTemplates = []courseDef{
	// 0 — Go тіліне кіріспе (Published)
	{
		title:  "Go тіліне кіріспе",
		desc:   "Go тілін нөлден үйреніңіз: типтер, параллельдік және стандартты кітапхана.",
		status: entities.StatusPublished,
		mods: [3]modDef{
			{title: "Go негіздері", lessons: [4]lessonDef{
				{title: "Go тіліне кіріспе", typ: entities.LessonText,
					content: "Go — Google әзірлеген статикалық типтелген компиляцияланатын тіл. Ол қарапайымдылық, тиімділік және кірістірілген параллельдік үшін жасалған. Бүгін Go-ны қолданатын үлкен компаниялар: Docker, Kubernetes, Cloudflare.",
					quiz: []qDef{
						{q: "Go тілін кім жасады?", correct: "Google", wrong: [3]string{"Microsoft", "Meta", "Apple"}},
						{q: "Go-да горутинді іске қосатын кілт сөз:", correct: "go", wrong: [3]string{"goroutine", "async", "thread"}},
					}},
				{title: "Орнату және баптау", typ: entities.LessonVideo,
					content: "Go-ны golang.org сайтынан жүктеп алыңыз. GOPATH және GOROOT айнымалыларын баптаңыз. VS Code-қа Go кеңейтімін орнатыңыз.",
					video:   "https://www.youtube.com/watch?v=Q0sKAMal4WQ"},
				{title: "Деректер типтері", typ: entities.LessonText,
					content: "Go-да негізгі типтер: int, float64, string, bool. Составные типтер: array, slice, map, struct. Типтік жүйе қатаң — айқын түрлендіру талап етіледі."},
				{title: "Басқару ағындары", typ: entities.LessonText,
					content: "Go-да if/else, for (жалғыз цикл операторы), switch операторлары бар. range итераторы slice, map және channel бойынша итерациялайды."},
			}},
			{title: "Функциялар мен интерфейстер", lessons: [4]lessonDef{
				{title: "Функциялар", typ: entities.LessonText,
					content: "Go функциялары бірнеше мән қайтара алады. Вариативті функциялар ... операторын қолданады. Анонимді функциялар мен жабылымдар да мүмкін.",
					quiz: []qDef{
						{q: "Go-да функциядан бірнеше мән қайтару мүмкін бе?", correct: "Иә, мүмкін", wrong: [3]string{"Жоқ", "Тек pointer арқылы", "Тек struct арқылы"}},
						{q: "Defer операторы қашан орындалады?", correct: "Функция аяқталғанда", wrong: [3]string{"Бірден", "Цикл соңында", "Горутин аяқталғанда"}},
					}},
				{title: "Пакеттер мен модульдер", typ: entities.LessonText,
					content: "Go модульдері go.mod файлымен анықталады. go get командасы тәуелділіктерді жүктейді. Пакеттер бір каталогтағы файлдар жиынтығы."},
				{title: "Интерфейстер", typ: entities.LessonVideo,
					content: "Go интерфейстері әдістер жиынтығын анықтайды. Имплиситті имплементация — ешқандай implements кілт сөзі жоқ.",
					video:   "https://www.youtube.com/watch?v=lh_Uv2imp14"},
				{title: "Тапсырмалар жинағы", typ: entities.LessonPDF,
					content: "Функциялар мен интерфейстерге арналған тапсырмалар.",
					file:    "https://example.com/go-tasks1.pdf"},
			}},
			{title: "Параллельдік", lessons: [4]lessonDef{
				{title: "Горутиндер", typ: entities.LessonText,
					content: "Горутиндер — Go жұмыс уақытымен басқарылатын жеңіл ағындар. go f() синтаксисімен іске қосылады. Олар OS ағындарынан бірнеше есе жеңіл — бастапқы стек тек 2KB.",
					quiz: []qDef{
						{q: "Горутиннің бастапқы стек өлшемі:", correct: "2KB", wrong: [3]string{"1MB", "64KB", "512KB"}},
						{q: "Каналға деректер жіберу операторы:", correct: "<-", wrong: [3]string{"->", ">>", "=>"}},
					}},
				{title: "Каналдар", typ: entities.LessonVideo,
					content: "Каналдар горутиндер арасындағы байланысты қамтамасыз етеді. make(chan T) буферсіз канал жасайды.",
					video:   "https://www.youtube.com/watch?v=oV9rvDllKEg"},
				{title: "sync пакеті", typ: entities.LessonText,
					content: "sync.WaitGroup горутиндерді синхрондауға мүмкіндік береді. sync.Mutex бөлісілген деректерді қорғайды."},
				{title: "Параллельдік ресурстары", typ: entities.LessonLink,
					content: "Go параллельдігі туралы ресми блог жазбасы.",
					link:    "https://go.dev/blog/pipelines"},
			}},
		},
		assigns: [2]aDef{
			{title: "CLI калькулятор жазу", desc: "Go тілінде командалық жолдан жұмыс істейтін калькулятор жазыңыз. Ол +, -, *, / операцияларын қолдауы керек.", pts: 100, days: 14},
			{title: "REST API жасау", desc: "Gin фреймворкімен кітаптарды басқаратын REST API жасаңыз. CRUD операциялары мен қате өңдеу міндетті.", pts: 100, days: 28},
		},
		exams: [2]eDef{
			{title: "Аралық бақылау", desc: "Go тілінің негіздерін бағалайтын аралық бақылау.", mins: 60, maxAttempts: 2, pts: 30, qs: []qDef{
				{q: "Go-да main функциясы қай пакетте орналасуы керек?", correct: "main", wrong: [3]string{"base", "root", "program"}},
				{q: "Nil pointer dereference не шығарады?", correct: "Panic", wrong: [3]string{"Error", "Warning", "nil"}},
				{q: "go vet командасы не үшін қолданылады?", correct: "Кодтағы қателерді табу", wrong: [3]string{"Тесттерді іске қосу", "Кодты компиляциялау", "Тәуелділіктерді жүктеу"}},
			}},
			{title: "Қорытынды емтихан", desc: "Go тілі бойынша барлық тақырыптарды қамтитын қорытынды емтихан.", mins: 90, maxAttempts: 1, pts: 50, qs: []qDef{
				{q: "sync.WaitGroup-тің Done() әдісі не істейді?", correct: "Санауышты бір азайтады", wrong: [3]string{"Горутинді тоқтатады", "Каналды жабады", "Санауышты нөлге қояды"}},
				{q: "Go-да интерфейсті имплементациялау тәсілі:", correct: "Имплиситті (кілт сөзсіз)", wrong: [3]string{"implements кілт сөзімен", "extends кілт сөзімен", "interface{} арқылы"}},
				{q: "Buffered channel не береді?", correct: "Буфер толғанша блокталмайды", wrong: [3]string{"Деректер жоғалмайды", "Жылдамырақ жұмыс", "Синхронды байланыс"}},
			}},
		},
		anns: []annDef{
			{title: "Курсқа қош келдіңіз!", body: "Бүгіннен бастап Go тілін бірге үйренеміз. Аптасына 2 сабақ жоспарланған. Сұрақтарыңызды форумда қалдырыңыз."},
			{title: "1-модуль материалдары жарияланды", body: "Go негіздері бойынша барлық дәріс материалдары порталда қол жетімді. Мұқият оқып шығыңыздар."},
			{title: "Аралық бақылау туралы хабарлама", body: "Аралық бақылау 3 апта ішінде өткізіледі. 3 сұрақ, 60 минут. Жақсы дайындалыңыз!"},
			{title: "Тапсырма мерзімі туралы еске салу", body: "CLI калькулятор тапсырмасының мерзімі — осы жұманың жұмасы. Кеш тапсырмалар 10% ұпай шегеруімен қабылданады."},
		},
	},

	// 1 — Деректер қорлары мен SQL (Published)
	{
		title:  "Деректер қорлары мен SQL",
		desc:   "Реляциялық дерекқорлар, SQL сұраулары, индекстер және PostgreSQL-мен жұмыс.",
		status: entities.StatusPublished,
		mods: [3]modDef{
			{title: "SQL негіздері", lessons: [4]lessonDef{
				{title: "Реляциялық модель", typ: entities.LessonText,
					content: "Реляциялық дерекқор деректерді кестелерде (реляцияларда) сақтайды. Әрбір кесте бағандар (атрибуттар) мен жолдардан (кортеждерден) тұрады. Бастапқы кілт (Primary Key) жолды бірегей анықтайды.",
					quiz: []qDef{
						{q: "SQL-де деректерді іріктеу операторы:", correct: "SELECT", wrong: [3]string{"FETCH", "GET", "READ"}},
						{q: "Кестедегі бірегей жолды анықтайтын кілт:", correct: "Primary Key", wrong: [3]string{"Foreign Key", "Index", "Unique Key"}},
					}},
				{title: "CREATE, INSERT, UPDATE, DELETE", typ: entities.LessonVideo,
					content: "DDL (Data Definition Language) және DML (Data Manipulation Language) операторлары туралы.",
					video:   "https://www.youtube.com/watch?v=HXV3zeQKqGY"},
				{title: "WHERE және шарттар", typ: entities.LessonText,
					content: "WHERE шарты сұраулардың нәтижесін сүзеді. AND, OR, NOT логикалық операторлары, LIKE, IN, BETWEEN, IS NULL шарттары қолданылады."},
				{title: "SQL негіздері: тапсырмалар", typ: entities.LessonText,
					content: "SELECT, INSERT, UPDATE, DELETE операторларын пайдаланып кестелерді басқару тапсырмалары."},
			}},
			{title: "JOIN және агрегация", lessons: [4]lessonDef{
				{title: "JOIN операциялары", typ: entities.LessonText,
					content: "INNER JOIN екі кестеден сәйкес жолдарды қайтарады. LEFT JOIN сол кестенің барлық жолдарын, RIGHT JOIN оң кестенің барлық жолдарын қайтарады. FULL OUTER JOIN екеуінің де барлық жолдарын қайтарады.",
					quiz: []qDef{
						{q: "INNER JOIN нені қайтарады?", correct: "Екі кестеден сәйкес жолдарды", wrong: [3]string{"Тек сол кестені", "Барлық жолдарды", "NULL мәндерін"}},
						{q: "GROUP BY нені орындайды?", correct: "Жолдарды топтастырады", wrong: [3]string{"Ретке келтіреді", "Сүзеді", "Біріктіреді"}},
					}},
				{title: "Агрегат функциялар", typ: entities.LessonText,
					content: "COUNT(), SUM(), AVG(), MIN(), MAX() агрегат функциялары деректерді топтастырып есептейді. GROUP BY топтастыру үшін, HAVING топтастырылған нәтижені сүзу үшін қолданылады."},
				{title: "Кіші сұраулар", typ: entities.LessonVideo,
					content: "Кіші сұраулар (subqueries) басқа сұраудың ішінде орналасқан SELECT операторлары.",
					video:   "https://www.youtube.com/watch?v=XasgLX78gGo"},
				{title: "JOIN тапсырмалары", typ: entities.LessonPDF,
					content: "JOIN операциялары мен агрегат функцияларына арналған тапсырмалар жинағы.",
					file:    "https://example.com/sql-joins.pdf"},
			}},
			{title: "Индекстер мен оңтайландыру", lessons: [4]lessonDef{
				{title: "Индекстер", typ: entities.LessonText,
					content: "Индекс — дерекқор кестесіндегі деректерді жылдам іздеу үшін қолданылатын деректер құрылымы. B-tree индексі ең жиі қолданылады. CREATE INDEX операторымен жасалады.",
					quiz: []qDef{
						{q: "Индекстің негізгі мақсаты:", correct: "Сұрауларды жылдамдату", wrong: [3]string{"Деректерді сақтау", "Кестені қорғау", "Резервтік көшірме жасау"}},
						{q: "EXPLAIN операторы не береді?", correct: "Сұраудың орындалу жоспарын", wrong: [3]string{"Кесте схемасын", "Индекстер тізімін", "Статистиканы"}},
					}},
				{title: "PostgreSQL ерекшеліктері", typ: entities.LessonVideo,
					content: "PostgreSQL — кеңейтілген мүмкіндіктері бар ашық бастапқы кодты RDBMS. JSON типі, массивтер, pgvector кеңейтімі.",
					video:   "https://www.youtube.com/watch?v=qw--VYLpxG4"},
				{title: "Транзакциялар", typ: entities.LessonText,
					content: "Транзакция — атомдық орындалатын SQL операторларының тізбегі. ACID қасиеттері: Atomicity, Consistency, Isolation, Durability."},
				{title: "PostgreSQL қосымша ресурстары", typ: entities.LessonLink,
					content: "PostgreSQL ресми құжаттамасы.",
					link:    "https://www.postgresql.org/docs/"},
			}},
		},
		assigns: [2]aDef{
			{title: "Дерекқор схемасын жобалау", desc: "Университет ақпараттық жүйесі үшін ER-диаграмма және SQL схемасын жасаңыз. Кем дегенде 5 кесте болуы керек.", pts: 100, days: 12},
			{title: "Күрделі SQL сұраулары", desc: "Берілген дерекқор схемасы бойынша JOIN, агрегация және кіші сұраулар қолданылатын 10 сұрау жазыңыз.", pts: 80, days: 25},
		},
		exams: [2]eDef{
			{title: "SQL Аралық бақылау", desc: "SQL негіздері мен JOIN операцияларын бағалайтын аралық бақылау.", mins: 60, maxAttempts: 2, pts: 30, qs: []qDef{
				{q: "LEFT JOIN-да сол кестеде сәйкестік болмаса не қайтарылады?", correct: "NULL", wrong: [3]string{"Қателік", "Бос жол", "0"}},
				{q: "DISTINCT кілт сөзі не істейді?", correct: "Қайталанатын жолдарды жояды", wrong: [3]string{"Ретке келтіреді", "Топтастырады", "Санайды"}},
				{q: "Индексті қай жағдайда қолдану тиімді?", correct: "Жиі оқылатын бағандарда", wrong: [3]string{"Барлық бағандарда", "Тек бастапқы кілтте", "Жиі жазылатын бағандарда"}},
			}},
			{title: "SQL Қорытынды емтихан", desc: "Барлық SQL тақырыптары бойынша қорытынды емтихан.", mins: 90, maxAttempts: 1, pts: 50, qs: []qDef{
				{q: "HAVING шарты немен ерекшеленеді?", correct: "Топтастырылған нәтижені сүзеді", wrong: [3]string{"WHERE-ден жылдамырақ", "Индексті қолданады", "JOIN-мен жұмыс істейді"}},
				{q: "Транзакциядағы ROLLBACK не істейді?", correct: "Өзгерістерді болдырмайды", wrong: [3]string{"Деректерді сақтайды", "Транзакцияны аяқтайды", "Кестені тазалайды"}},
				{q: "pgvector кеңейтімі не үшін қолданылады?", correct: "Векторлық іздеу", wrong: [3]string{"Мәтіндік іздеу", "Шифрлеу", "Репликация"}},
			}},
		},
		anns: []annDef{
			{title: "Курсқа қош келдіңіз!", body: "SQL және PostgreSQL бойынша курсты бастаймыз. Жаттығу үшін бізге берілген тестілік дерекқорды қолданыңыз."},
			{title: "PostgreSQL орнату нұсқаулығы", body: "postgresql.org сайтынан PostgreSQL 16 нұсқасын жүктеп алыңыз. pgAdmin құралын да орнатуды ұсынамыз."},
			{title: "Бірінші тапсырма жарияланды", body: "Дерекқор схемасын жобалау тапсырмасы порталда. Мерзімі — 12 күн."},
		},
	},

	// 2 — React және TypeScript (Published)
	{
		title:  "React және TypeScript",
		desc:   "Заманауи фронтенд әзірлеу: React хуктары, TypeScript типтері және TanStack Query.",
		status: entities.StatusPublished,
		mods: [3]modDef{
			{title: "React негіздері", lessons: [4]lessonDef{
				{title: "Компоненттер мен JSX", typ: entities.LessonText,
					content: "React компоненттері — UI-дің қайта қолданылатын бөліктері. JSX — JavaScript-те HTML жазуға мүмкіндік беретін синтаксистік қант. Компоненттер функционалды немесе класс негізінде болады.",
					quiz: []qDef{
						{q: "React-та компоненттің күйін сақтайтын хук:", correct: "useState", wrong: [3]string{"useRef", "useMemo", "useContext"}},
						{q: "JSX-та className не үшін қолданылады?", correct: "CSS класстарын қосу үшін", wrong: [3]string{"class JavaScript-тің кілт сөзі болғандықтан", "Жылдамдық үшін", "TypeScript талабы"}},
					}},
				{title: "Props пен State", typ: entities.LessonVideo,
					content: "Props — компоненттер арасындағы деректер ағыны. State — компоненттің ішкі күйі.",
					video:   "https://www.youtube.com/watch?v=4UZrsTqkcW4"},
				{title: "Тізімдер мен кілттер", typ: entities.LessonText,
					content: "React-та тізімдерді map() арқылы рендерлейміз. Әрбір элементтің бірегей key prop-ы болуы керек."},
				{title: "Оқиғаларды өңдеу", typ: entities.LessonText,
					content: "React оқиғалары синтетикалық оқиғалар ретінде ұсынылады. onClick, onChange, onSubmit т.б. пайдаланыңыз."},
			}},
			{title: "TypeScript пен React", lessons: [4]lessonDef{
				{title: "TypeScript типтері", typ: entities.LessonText,
					content: "TypeScript JavaScript-ке статикалық типтерді қосады. interface, type, union types, generics. React компоненттерінде props типтерін анықтаңыз.",
					quiz: []qDef{
						{q: "React компонентінің props типін анықтауға арналған TypeScript конструкциясы:", correct: "interface немесе type", wrong: [3]string{"class", "enum", "namespace"}},
						{q: "TypeScript-тегі optional prop белгісі:", correct: "?", wrong: [3]string{"!", "&", "|"}},
					}},
				{title: "Generics пен utility types", typ: entities.LessonText,
					content: "Generic типтер қайта қолданылатын компоненттер жасауға мүмкіндік береді. Partial<T>, Required<T>, Pick<T,K>, Omit<T,K> utility типтері."},
				{title: "useEffect хуку", typ: entities.LessonVideo,
					content: "useEffect — жанама эффекттерді (API сұраулар, subscriptions) өңдейді.",
					video:   "https://www.youtube.com/watch?v=0ZJgIjIuY7U"},
				{title: "TypeScript анықтамалығы", typ: entities.LessonPDF,
					content: "TypeScript Cheat Sheet — негізгі конструкциялар жинағы.",
					file:    "https://example.com/ts-cheatsheet.pdf"},
			}},
			{title: "Жай-күй басқарысы", lessons: [4]lessonDef{
				{title: "Zustand пен TanStack Query", typ: entities.LessonText,
					content: "Zustand — минималистік клиенттік state management кітапханасы. TanStack Query — сервер күйін кэштейді, синхрондайды. Бұл екеуі толықтырады бір-бірін.",
					quiz: []qDef{
						{q: "TanStack Query-дің негізгі артықшылығы:", correct: "Сервер деректерін кэштеу", wrong: [3]string{"DOM-ды жаңарту", "TypeScript қолдауы", "Маршруттау"}},
						{q: "Zustand дегеніміз не?", correct: "State management кітапханасы", wrong: [3]string{"HTTP клиент", "UI компонент кітапханасы", "Тестілеу фреймворкі"}},
					}},
				{title: "Context API", typ: entities.LessonVideo,
					content: "React Context API — prop drilling-ті болдырмайды. createContext, Provider, useContext.",
					video:   "https://www.youtube.com/watch?v=HYKDUF8X3qI"},
				{title: "Өнімділікті оңтайландыру", typ: entities.LessonText,
					content: "React.memo, useMemo, useCallback — қайта рендерлеуді азайтады. Lazy loading және Suspense."},
				{title: "React ресурстары", typ: entities.LessonLink,
					content: "React ресми құжаттамасы.",
					link:    "https://react.dev"},
			}},
		},
		assigns: [2]aDef{
			{title: "Todo қосымшасы", desc: "React пен TypeScript қолданып толыққанды Todo қосымшасын жасаңыз. CRUD операциялары, localStorage, TypeScript типтері міндетті.", pts: 100, days: 14},
			{title: "Dashboard компоненті", desc: "TanStack Query мен Zustand қолданып деректерді нақты уақытта жаңартатын dashboard жасаңыз.", pts: 120, days: 30},
		},
		exams: [2]eDef{
			{title: "React Аралық бақылау", desc: "React хуктары мен компоненттері бойынша аралық бақылау.", mins: 45, maxAttempts: 2, pts: 30, qs: []qDef{
				{q: "useEffect-тегі бос dependency array [] нені білдіреді?", correct: "Тек бір рет іске қосылады", wrong: [3]string{"Ешқашан іске қосылмайды", "Күнде іске қосылады", "Қателік тудырады"}},
				{q: "React-та key prop неге керек?", correct: "Тізім элементтерін бірегей анықтауға", wrong: [3]string{"CSS стильдеу үшін", "TypeScript талабы", "Props беру үшін"}},
				{q: "useState хуку нені қайтарады?", correct: "[мән, setter функциясы]", wrong: [3]string{"Тек мән", "Тек setter", "Object {get, set}"}},
			}},
			{title: "React Қорытынды емтихан", desc: "React және TypeScript бойынша толық қорытынды емтихан.", mins: 75, maxAttempts: 1, pts: 50, qs: []qDef{
				{q: "Zustand store-ын жасайтын функция:", correct: "create()", wrong: [3]string{"useState()", "useStore()", "createStore()"}},
				{q: "TypeScript-те Record<string, number> нені білдіреді?", correct: "string кілт пен number мән сөздігі", wrong: [3]string{"number массиві", "string массиві", "Кез келген объект"}},
				{q: "React.memo не үшін қолданылады?", correct: "Пропстар өзгермесе қайта рендерлеуді болдырмайды", wrong: [3]string{"Компонентті жадта сақтайды", "Асинхронды жұмыс үшін", "TypeScript интеграциясы"}},
			}},
		},
		anns: []annDef{
			{title: "React 19 жаңалықтары", body: "Курста React 19 нұсқасы қолданылады. Жаңа Server Components және Actions мүмкіндіктерін бірге зерттейміз."},
			{title: "Node.js орнату", body: "Курсты бастамас бұрын Node.js 20+ нұсқасын орнатыңыз. npm пакет менеджерін де тексеріп алыңыз."},
			{title: "Todo тапсырмасы жарияланды", body: "Бірінші тапсырма — Todo қосымшасы. Мерзімі 2 апта. TypeScript қателерсіз болуы керек."},
		},
	},

	// 3 — Машиналық оқыту негіздері (Published)
	{
		title:  "Машиналық оқыту негіздері",
		desc:   "МО негіздері: бақылаулы оқыту, нейрондық желілер және Python экожүйесі.",
		status: entities.StatusPublished,
		mods: [3]modDef{
			{title: "МО кіріспе", lessons: [4]lessonDef{
				{title: "Машиналық оқыту дегеніміз не?", typ: entities.LessonText,
					content: "Машиналық оқыту — компьютерлерге деректерден үйренуге мүмкіндік беретін алгоритмдер жиынтығы. Бақылаулы, бақылаусыз және күшейту арқылы оқыту — негізгі үш парадигма.",
					quiz: []qDef{
						{q: "Бақылаулы оқытудың негізгі белгісі:", correct: "Таңбаланған деректер қолданылады", wrong: [3]string{"Ережелер жазылады", "Деректер жоқ", "Нейрондық желі міндетті"}},
						{q: "Keras/TensorFlow-да модельді компиляциялайтын әдіс:", correct: "model.compile()", wrong: [3]string{"model.build()", "model.train()", "model.fit()"}},
					}},
				{title: "Python экожүйесі", typ: entities.LessonVideo,
					content: "NumPy, Pandas, Matplotlib, Scikit-learn — МО үшін негізгі Python кітапханалары.",
					video:   "https://www.youtube.com/watch?v=aircAruvnKk"},
				{title: "Деректерді дайындау", typ: entities.LessonText,
					content: "Деректерді тазалау, нормализация, кодтау. Train/validation/test бөліктеу. sklearn.preprocessing пен sklearn.model_selection."},
				{title: "МО негіздері: тапсырмалар", typ: entities.LessonText,
					content: "Python пен Pandas арқылы деректерді зерттеу тапсырмалары."},
			}},
			{title: "Бақылаулы оқыту алгоритмдері", lessons: [4]lessonDef{
				{title: "Регрессия мен жіктеу", typ: entities.LessonText,
					content: "Сызықтық регрессия үздіксіз мәндерді болжайды. Логистикалық регрессия екілік жіктеу үшін. Random Forest ансамблді әдіс.",
					quiz: []qDef{
						{q: "Регрессия мен жіктеудің айырмашылығы:", correct: "Регрессия — сан, жіктеу — категория", wrong: [3]string{"Ешқандай айырмашылық жоқ", "Регрессия жылдамырақ", "Жіктеу дәлірек"}},
						{q: "Overfitting нені білдіреді?", correct: "Модель тренинг деректерін жаттайды", wrong: [3]string{"Модель дұрыс жұмыс істейді", "Деректер аз", "Алгоритм дұрыс таңдалмаған"}},
					}},
				{title: "Decision Tree мен Random Forest", typ: entities.LessonText,
					content: "Decision Tree деректерді ағаш тәрізді шешімдер арқылы бөледі. Random Forest — бірнеше ағаш ансамблі."},
				{title: "Модельді бағалау", typ: entities.LessonVideo,
					content: "Accuracy, Precision, Recall, F1-score, AUC-ROC — жіктеу метрикалары.",
					video:   "https://www.youtube.com/watch?v=85dtiMz9tSo"},
				{title: "Scikit-learn анықтамалығы", typ: entities.LessonPDF,
					content: "Scikit-learn API анықтамалығы PDF форматында.",
					file:    "https://example.com/sklearn-ref.pdf"},
			}},
			{title: "Нейрондық желілер", lessons: [4]lessonDef{
				{title: "Нейрондық желілер негіздері", typ: entities.LessonText,
					content: "Жасанды нейрондық желі — адам миының математикалық моделі. Қабаттар: input, hidden, output. Активация функциялары: ReLU, sigmoid, softmax.",
					quiz: []qDef{
						{q: "Hidden layer неліктен керек?", correct: "Сызықты емес мүмкіндіктерді үйрену үшін", wrong: [3]string{"Жылдамдату үшін", "Деректерді сақтау үшін", "Нәтижені шығару үшін"}},
						{q: "Backpropagation не істейді?", correct: "Градиент арқылы салмақтарды жаңартады", wrong: [3]string{"Деректерді нормализациялайды", "Модельді сақтайды", "Болжам жасайды"}},
					}},
				{title: "Keras арқылы МО", typ: entities.LessonVideo,
					content: "TensorFlow/Keras пайдаланып нейрондық желі жасау.",
					video:   "https://www.youtube.com/watch?v=tPYj3fFJGjk"},
				{title: "Dropout пен Regularization", typ: entities.LessonText,
					content: "L1/L2 regularization және Dropout — overfitting-ті азайту тәсілдері. Batch Normalization жаттығу процесін тұрақтандырады."},
				{title: "МО ресурстары", typ: entities.LessonLink,
					content: "fast.ai курсы — тереңдетілген МО оқыту ресурсы.",
					link:    "https://fast.ai"},
			}},
		},
		assigns: [2]aDef{
			{title: "Жіктеу моделі жасау", desc: "Scikit-learn пайдаланып Titanic датасетіне жіктеу моделі жасаңыз. Нәтижені Accuracy және F1-score арқылы бағалаңыз.", pts: 100, days: 16},
			{title: "Нейрондық желі жасау", desc: "Keras пайдаланып MNIST датасетіндегі цифрларды тану нейрондық желісін жасаңыз. 95%+ accuracy мақсат.", pts: 150, days: 35},
		},
		exams: [2]eDef{
			{title: "МО Аралық бақылау", desc: "Машиналық оқытудың негізгі тұжырымдамаларын тексеретін аралық бақылау.", mins: 60, maxAttempts: 2, pts: 30, qs: []qDef{
				{q: "Cross-validation не береді?", correct: "Модельдің жалпыланушылығын бағалайды", wrong: [3]string{"Деректерді тазалайды", "Параметрлерді баптайды", "Нейрондық желі жасайды"}},
				{q: "Feature scaling не үшін керек?", correct: "Алгоритмдердің конвергенциясын жылдамдату", wrong: [3]string{"Деректерді сақтау", "Классты болжау", "Нәтижені визуализациялау"}},
				{q: "Confusion matrix нені көрсетеді?", correct: "TP, FP, TN, FN мәндерін", wrong: [3]string{"Шығын функциясын", "Оқыту жылдамдығын", "Деректер санын"}},
			}},
			{title: "МО Қорытынды емтихан", desc: "МО барлық тақырыптары бойынша қорытынды емтихан.", mins: 90, maxAttempts: 1, pts: 50, qs: []qDef{
				{q: "GAN нені білдіреді?", correct: "Generative Adversarial Network", wrong: [3]string{"Gradient Adaptive Network", "Graph Attention Node", "General AI Network"}},
				{q: "Transfer Learning-дің артықшылығы:", correct: "Аз деректермен жақсы нәтиже береді", wrong: [3]string{"Жылдам жұмыс істейді", "Кодты азайтады", "GPU қажет емес"}},
				{q: "Attention механизмі не үшін қолданылады?", correct: "Маңызды мүмкіндіктерге назар аудару", wrong: [3]string{"Деректерді сүзу", "Активация функциясы ретінде", "Regularization үшін"}},
			}},
		},
		anns: []annDef{
			{title: "Курсқа қош келдіңіз!", body: "МО курсын бастаймыз. Python 3.10+ пен Jupyter Notebook орнатыңыз. Google Colab-ты да пайдалануға болады."},
			{title: "Датасеттер туралы ақпарат", body: "Курста Kaggle датасеттері қолданылады. Тіркелу тегін. kaggle.com/datasets бетіне кіріңіз."},
			{title: "Бірінші тапсырма: Titanic", body: "Titanic датасетіне негізделген жіктеу тапсырмасы жарияланды. Jupyter Notebook форматында тапсырыңыз."},
			{title: "МО семинары хабарламасы", body: "Келесі жұма сағат 15:00-де онлайн семинар өткізіледі. Нейрондық желілер тақырыбы қаралады."},
		},
	},

	// 4 — Желілік қауіпсіздік (Published)
	{
		title:  "Желілік қауіпсіздік",
		desc:   "Желілік шабуылдар, қорғаныс механизмдері, шифрлеу және веб-қауіпсіздік.",
		status: entities.StatusPublished,
		mods: [3]modDef{
			{title: "Желі негіздері", lessons: [4]lessonDef{
				{title: "OSI моделі мен TCP/IP", typ: entities.LessonText,
					content: "OSI — 7 қабатты желілік модель: Физикалық, Деректер, Желілік, Тасымал, Сессия, Ұсыну, Қолданба. TCP/IP — практикалық 4 қабатты модель.",
					quiz: []qDef{
						{q: "OSI моделіндегі қабаттар саны:", correct: "7", wrong: [3]string{"4", "5", "3"}},
						{q: "TCP мен UDP айырмашылығы:", correct: "TCP сенімді, UDP жылдам", wrong: [3]string{"TCP жылдам, UDP сенімді", "Ешқандай айырмашылық жоқ", "TCP тек веб үшін"}},
					}},
				{title: "Желілік шабуыл түрлері", typ: entities.LessonVideo,
					content: "DDoS, Man-in-the-Middle, SQL Injection, XSS, CSRF — жиі кездесетін шабуылдар.",
					video:   "https://www.youtube.com/watch?v=hkKCBTXfJsA"},
				{title: "Брандмауэр мен IDS/IPS", typ: entities.LessonText,
					content: "Брандмауэр желілік трафикті сүзеді. IDS шабуылдарды анықтайды, IPS оларды блоктайды."},
				{title: "Желі қауіпсіздігі тапсырмалары", typ: entities.LessonText,
					content: "Wireshark пайдаланып желілік трафикті талдау тапсырмалары."},
			}},
			{title: "Шифрлеу", lessons: [4]lessonDef{
				{title: "Симметриялық және асимметриялық шифрлеу", typ: entities.LessonText,
					content: "Симметриялық шифрлеу: AES, DES — бір кілт. Асимметриялық: RSA, ECC — ашық/жабық кілт жұбы. HTTPS — TLS арқылы асимметриялық шифрлеу.",
					quiz: []qDef{
						{q: "HTTPS-те қолданылатын протокол:", correct: "TLS", wrong: [3]string{"SSL 2.0", "PGP", "AES"}},
						{q: "Hash функциясының қасиеті:", correct: "Бір бағытты (кері айналдыруға болмайды)", wrong: [3]string{"Екі бағытты", "Тез жұмыс істейді", "Кілт қолданады"}},
					}},
				{title: "PKI және сертификаттар", typ: entities.LessonText,
					content: "Public Key Infrastructure: CA, сертификаттар, CRL. X.509 сертификат форматы."},
				{title: "JWT және OAuth 2.0", typ: entities.LessonVideo,
					content: "JSON Web Token — аутентификация мен авторизация. OAuth 2.0 делегацияланған қол жеткізу.",
					video:   "https://www.youtube.com/watch?v=7Q17ubqLfaM"},
				{title: "Криптография анықтамалығы", typ: entities.LessonPDF,
					content: "Криптография тұжырымдамалары мен алгоритмдері анықтамалығы.",
					file:    "https://example.com/crypto-ref.pdf"},
			}},
			{title: "Веб-қауіпсіздік", lessons: [4]lessonDef{
				{title: "OWASP Top 10", typ: entities.LessonText,
					content: "OWASP Top 10 — веб-қосымшалардағы жиі кездесетін 10 қауіпсіздік тәуекелі. SQL Injection, XSS, CSRF, Broken Authentication т.б.",
					quiz: []qDef{
						{q: "SQL Injection-дан қорғану тәсілі:", correct: "Параметрленген сұраулар (Prepared Statements)", wrong: [3]string{"Деректерді шифрлеу", "HTTPS қолдану", "Firewall орнату"}},
						{q: "CSP (Content Security Policy) не үшін қолданылады?", correct: "XSS шабуылдарын болдырмау", wrong: [3]string{"CSRF қорғанысы", "SQL Injection", "DDoS қорғанысы"}},
					}},
				{title: "Аутентификация мен авторизация", typ: entities.LessonVideo,
					content: "MFA, SSO, RBAC — заманауи қауіпсіздік тәжірибелері.",
					video:   "https://www.youtube.com/watch?v=GhrvZ5nUWNg"},
				{title: "Penetration Testing негіздері", typ: entities.LessonText,
					content: "Pen testing кезеңдері: барлау, сканерлеу, эксплуатация, есеп беру. Этикалық хакинг."},
				{title: "Қауіпсіздік ресурстары", typ: entities.LessonLink,
					content: "OWASP ресми сайты.",
					link:    "https://owasp.org"},
			}},
		},
		assigns: [2]aDef{
			{title: "Қауіпсіздік аудиті", desc: "Берілген веб-қосымшаны OWASP Top 10 бойынша тексеріп, табылған осалдықтар туралы есеп жазыңыз.", pts: 100, days: 18},
			{title: "JWT аутентификация іске асыру", desc: "Go немесе Python тілінде JWT негізінде аутентификация жүйесін іске асырыңыз. Refresh token де болуы керек.", pts: 100, days: 30},
		},
		exams: [2]eDef{
			{title: "Қауіпсіздік Аралық бақылау", desc: "Желілік қауіпсіздік негіздері бойынша аралық бақылау.", mins: 60, maxAttempts: 2, pts: 30, qs: []qDef{
				{q: "Man-in-the-Middle шабуылынан қорғану:", correct: "TLS/HTTPS қолдану", wrong: [3]string{"Күшті пароль", "Firewall", "VPN ғана"}},
				{q: "bcrypt хэш функциясының артықшылығы:", correct: "Salt пен баяу алгоритм — brute force-қа төзімді", wrong: [3]string{"Жылдам жұмыс", "Кері айналдыруға болады", "Кілт қолданбайды"}},
				{q: "CSRF шабуылы нені пайдаланады?", correct: "Браузердегі аутентификация cookie-лерін", wrong: [3]string{"SQL осалдықтарды", "XSS-ті", "Ашық портты"}},
			}},
			{title: "Қауіпсіздік Қорытынды емтихан", desc: "Желілік қауіпсіздік пен криптография бойынша қорытынды емтихан.", mins: 90, maxAttempts: 1, pts: 50, qs: []qDef{
				{q: "Zero-day осалдығы дегеніміз:", correct: "Әзірлеуші хабарсыз осалдық", wrong: [3]string{"Ескі осалдық", "Патч жарияланған осалдық", "Теориялық осалдық"}},
				{q: "RSA шифрлеуінің математикалық негізі:", correct: "Үлкен сандарды жіктеудің күрделілігі", wrong: [3]string{"Дискретті логарифм", "Эллиптикалық қисық", "Hash коллизия"}},
				{q: "SOC (Security Operations Center) не үшін керек?", correct: "Қауіпсіздік оқиғаларын бақылау", wrong: [3]string{"Код жазу", "Жүйені баяулату", "Пайдаланушыларды басқару"}},
			}},
		},
		anns: []annDef{
			{title: "Курсқа қош келдіңіз!", body: "Желілік қауіпсіздік курсы — теория мен практиканың үйлесімі. Kali Linux VM орнатып алыңыз."},
			{title: "CTF жарысы туралы хабарлама", body: "Келесі ай ішінде онлайн CTF (Capture The Flag) жарысы өткізіледі. Тіркелу міндетті емес, бірақ ұсынылады."},
			{title: "Тапсырма жарияланды", body: "Қауіпсіздік аудиті тапсырмасы жарияланды. OWASP ZAP немесе Burp Suite Community Edition қолдануға болады."},
		},
	},

	// 5 — Бұлтты технологиялар (Published)
	{
		title:  "Бұлтты технологиялар",
		desc:   "Cloud computing негіздері, Docker, Kubernetes және CI/CD тәжірибелері.",
		status: entities.StatusPublished,
		mods: [3]modDef{
			{title: "Бұлтқа кіріспе", lessons: [4]lessonDef{
				{title: "Cloud Computing негіздері", typ: entities.LessonText,
					content: "Бұлтты есептеу — интернет арқылы ұсынылатын есептеу ресурстары. IaaS, PaaS, SaaS модельдері. AWS, Google Cloud, Azure — негізгі провайдерлер.",
					quiz: []qDef{
						{q: "IaaS дегеніміз:", correct: "Infrastructure as a Service", wrong: [3]string{"Internet as a Service", "Integration as a Service", "Input as a Service"}},
						{q: "Serverless архитектурасының артықшылығы:", correct: "Сервер басқарусыз код іске қосылады", wrong: [3]string{"Тегін", "Жылдамырақ", "Интернетсіз жұмыс"}},
					}},
				{title: "AWS негізгі сервистері", typ: entities.LessonVideo,
					content: "EC2, S3, RDS, Lambda, VPC — AWS-тің негізгі сервистері.",
					video:   "https://www.youtube.com/watch?v=a9__D53WsMs"},
				{title: "Бұлт қауіпсіздігі", typ: entities.LessonText,
					content: "IAM (Identity and Access Management), security groups, VPC, encryption at rest and in transit."},
				{title: "Cloud тапсырмалары", typ: entities.LessonText,
					content: "AWS Free Tier аккаунтын пайдаланып EC2 инстанс іске қосу тапсырмалары."},
			}},
			{title: "Docker мен контейнерлер", lessons: [4]lessonDef{
				{title: "Docker негіздері", typ: entities.LessonText,
					content: "Docker — контейнерлеу платформасы. Image — оқуға ғана арналған шаблон. Container — image-тің іске қосылған данасы. Dockerfile — image жасау нұсқаулары.",
					quiz: []qDef{
						{q: "Dockerfile-да ENTRYPOINT не анықтайды?", correct: "Контейнер іске қосылғандағы негізгі команда", wrong: [3]string{"Ортаны баптайды", "Файлдарды көшіреді", "Портты ашады"}},
						{q: "docker-compose не үшін қолданылады?", correct: "Бірнеше контейнерді бірге басқару", wrong: [3]string{"Image жасау", "Контейнерді мониторлеу", "Желіні баптау"}},
					}},
				{title: "Docker Compose", typ: entities.LessonVideo,
					content: "docker-compose.yml арқылы микросервистерді оркестрациялау.",
					video:   "https://www.youtube.com/watch?v=DM65_JyGxCo"},
				{title: "Docker Registry", typ: entities.LessonText,
					content: "Docker Hub, AWS ECR, GitHub Container Registry. docker push/pull командалары."},
				{title: "Docker анықтамалығы", typ: entities.LessonPDF,
					content: "Docker команда жолы анықтамалығы.",
					file:    "https://example.com/docker-cheat.pdf"},
			}},
			{title: "Kubernetes", lessons: [4]lessonDef{
				{title: "Kubernetes негіздері", typ: entities.LessonText,
					content: "Kubernetes (K8s) — контейнерлерді автоматты оркестрациялайды. Pod — ең кіші орналастыру бірлігі. Deployment — подтардың үлгісін анықтайды.",
					quiz: []qDef{
						{q: "Kubernetes-тегі ең кіші орналастыру бірлігі:", correct: "Pod", wrong: [3]string{"Container", "Node", "Cluster"}},
						{q: "K8s Service нені атқарады?", correct: "Подтарға тұрақты желілік қол жеткізуді қамтамасыз етеді", wrong: [3]string{"Деректерді сақтайды", "Кодты іске қосады", "Логтарды жинайды"}},
					}},
				{title: "K8s Deployments мен Services", typ: entities.LessonVideo,
					content: "kubectl командалары, YAML манифесттер, rolling updates.",
					video:   "https://www.youtube.com/watch?v=X48VuDVv0do"},
				{title: "Helm Charts", typ: entities.LessonText,
					content: "Helm — Kubernetes үшін пакет менеджері. Chart — K8s ресурстарының шаблоны."},
				{title: "K8s ресурстары", typ: entities.LessonLink,
					content: "Kubernetes ресми оқулығы.",
					link:    "https://kubernetes.io/docs/tutorials/"},
			}},
		},
		assigns: [2]aDef{
			{title: "Docker-де қосымша орналастыру", desc: "Go немесе Python қосымшасын Dockerize жасаңыз. Multi-stage build қолданыңыз. docker-compose.yml жасаңыз.", pts: 100, days: 15},
			{title: "K8s кластерін баптау", desc: "Minikube немесе kind қолданып локалды K8s кластерін баптаңыз. Қосымшаны Deployment пен Service арқылы орналастырыңыз.", pts: 120, days: 28},
		},
		exams: [2]eDef{
			{title: "Бұлт Аралық бақылау", desc: "Docker мен бұлтты технологиялар бойынша аралық бақылау.", mins: 60, maxAttempts: 2, pts: 30, qs: []qDef{
				{q: "Docker image пен container айырмашылығы:", correct: "Image — шаблон, container — іске қосылған дана", wrong: [3]string{"Ешқандай айырмашылық жоқ", "Container тезірек", "Image үлкенірек"}},
				{q: "K8s ReplicaSet не береді?", correct: "Подтардың белгілі санын қамтамасыз етеді", wrong: [3]string{"Желіні баптайды", "Деректерді сақтайды", "Логтарды жинайды"}},
				{q: "CI/CD-да CD нені білдіреді?", correct: "Continuous Delivery/Deployment", wrong: [3]string{"Code Development", "Container Deployment", "Cloud Distribution"}},
			}},
			{title: "Бұлт Қорытынды емтихан", desc: "Бұлтты технологиялар мен K8s бойынша қорытынды емтихан.", mins: 90, maxAttempts: 1, pts: 50, qs: []qDef{
				{q: "K8s Ingress нені атқарады?", correct: "Сыртқы HTTP трафикті ішкі сервистерге бағыттайды", wrong: [3]string{"Деректерді шифрлейді", "Подтарды жасайды", "Логтарды сақтайды"}},
				{q: "GitOps принципі:", correct: "Git репозиторий — инфрақұрылым конфигурациясының бірден-бір дерек көзі", wrong: [3]string{"Git арқылы CI/CD баптау", "GitHub Actions қолдану", "Git hooks пайдалану"}},
				{q: "Service Mesh нені шешеді?", correct: "Микросервистер арасындағы байланыс күрделілігін", wrong: [3]string{"Деректерді сақтау", "Frontend маршруттау", "Код тестілеу"}},
			}},
		},
		anns: []annDef{
			{title: "Курсқа қош келдіңіз!", body: "Бұлтты технологиялар курсын бастаймыз. Docker Desktop және AWS Free Tier аккаунтын алдын ала орнатып алыңыздар."},
			{title: "Docker Desktop баламасы", body: "MacOS-та Docker Desktop баяу болса, OrbStack пайдалануды ұсынамыз. Ол жылдам әрі ресурс тиімді."},
			{title: "K8s тапсырмасы жарияланды", body: "Minikube тапсырмасы порталда. Kubectl командаларын жаттығу үшін killercoda.com пайдаланыңыз."},
		},
	},

	// 6 — Мобильді қосымшалар (Draft)
	{
		title:  "Мобильді қосымшалар",
		desc:   "React Native пен Expo арқылы iOS және Android қосымшаларын жасау.",
		status: entities.StatusDraft,
		mods: [3]modDef{
			{title: "React Native негіздері", lessons: [4]lessonDef{
				{title: "React Native кіріспе", typ: entities.LessonText,
					content: "React Native — React компоненттерін пайдаланып нативті мобильді қосымшалар жасауға мүмкіндік береді."},
				{title: "Expo баптау", typ: entities.LessonVideo,
					content: "Expo — React Native әзірлеуді жеңілдететін платформа.",
					video:   "https://www.youtube.com/watch?v=0-S5a0eXPoc"},
				{title: "Негізгі компоненттер", typ: entities.LessonText,
					content: "View, Text, Image, ScrollView, TextInput — React Native-тің негізгі компоненттері."},
				{title: "StyleSheet API", typ: entities.LessonText,
					content: "React Native-та стильдеу CSS-ке ұқсас, бірақ объект негізінде."},
			}},
			{title: "Навигация мен жай-күй", lessons: [4]lessonDef{
				{title: "React Navigation", typ: entities.LessonText,
					content: "Stack, Tab, Drawer навигаторлары. react-navigation кітапханасы."},
				{title: "Async Storage", typ: entities.LessonText,
					content: "AsyncStorage — мобильді құрылғыда деректерді сақтауға арналған жергілікті қойма."},
				{title: "Redux Toolkit", typ: entities.LessonVideo,
					content: "Redux Toolkit арқылы мобильді қосымшада күй басқарысы.",
					video:   "https://www.youtube.com/watch?v=bbkBuqC1rU4"},
				{title: "Жай-күй тапсырмалары", typ: entities.LessonPDF,
					content: "Navigation мен State тапсырмалары.",
					file:    "https://example.com/rn-tasks.pdf"},
			}},
			{title: "API мен жарияланым", lessons: [4]lessonDef{
				{title: "REST API байланысы", typ: entities.LessonText,
					content: "fetch API немесе axios кітапханасы арқылы бэкендпен байланыс."},
				{title: "Push Notifications", typ: entities.LessonVideo,
					content: "Expo Notifications арқылы push хабарландырулар жіберу.",
					video:   "https://www.youtube.com/watch?v=8NKE_fGiOqA"},
				{title: "App Store & Play Store", typ: entities.LessonText,
					content: "EAS Build арқылы қосымшаны App Store пен Google Play-ге жариялау."},
				{title: "React Native ресурстары", typ: entities.LessonLink,
					content: "React Native ресми құжаттамасы.",
					link:    "https://reactnative.dev"},
			}},
		},
	},

	// 7 — DevOps практикалары (Draft)
	{
		title:  "DevOps практикалары",
		desc:   "CI/CD, Infrastructure as Code, мониторинг және SRE тәжірибелері.",
		status: entities.StatusDraft,
		mods: [3]modDef{
			{title: "CI/CD негіздері", lessons: [4]lessonDef{
				{title: "DevOps мәдениеті", typ: entities.LessonText,
					content: "DevOps — әзірлеу мен операциялар командаларының үздіксіз ынтымақтастығы. CI/CD, IaC, мониторинг — негізгі тәжірибелер."},
				{title: "GitHub Actions", typ: entities.LessonVideo,
					content: "GitHub Actions арқылы CI/CD pipeline жасау.",
					video:   "https://www.youtube.com/watch?v=R8_veQiYBjI"},
				{title: "GitLab CI/CD", typ: entities.LessonText,
					content: ".gitlab-ci.yml файлы арқылы pipeline баптау. Stages, jobs, artifacts."},
				{title: "CI/CD тапсырмалары", typ: entities.LessonText,
					content: "GitHub Actions арқылы Go қосымшасына CI pipeline жасау тапсырмасы."},
			}},
			{title: "Infrastructure as Code", lessons: [4]lessonDef{
				{title: "Terraform негіздері", typ: entities.LessonText,
					content: "Terraform — HashiCorp-тың IaC құралы. HCL тілінде инфрақұрылым сипатталады. Plan → Apply → Destroy."},
				{title: "Ansible", typ: entities.LessonText,
					content: "Ansible — конфигурация менеджері. YAML playbooks арқылы серверлерді баптайды. Agentless."},
				{title: "Terraform тапсырмасы", typ: entities.LessonVideo,
					content: "Terraform арқылы AWS-те ресурстар жасау.",
					video:   "https://www.youtube.com/watch?v=SLB_c_ayRMo"},
				{title: "IaC анықтамалығы", typ: entities.LessonPDF,
					content: "Terraform командалар анықтамалығы.",
					file:    "https://example.com/terraform-cheat.pdf"},
			}},
			{title: "Мониторинг пен журналдау", lessons: [4]lessonDef{
				{title: "Prometheus мен Grafana", typ: entities.LessonText,
					content: "Prometheus метрикаларды жинайды. Grafana визуализация үшін. PromQL сұрау тілі."},
				{title: "ELK Stack", typ: entities.LessonVideo,
					content: "Elasticsearch, Logstash, Kibana — журналдарды жинау мен талдау.",
					video:   "https://www.youtube.com/watch?v=MRMgd6E9AXE"},
				{title: "SRE тәжірибелері", typ: entities.LessonText,
					content: "SLI, SLO, SLA метрикалары. Error budget. On-call тәжірибесі."},
				{title: "DevOps ресурстары", typ: entities.LessonLink,
					content: "Google SRE кітабы онлайн нұсқасы.",
					link:    "https://sre.google/books/"},
			}},
		},
	},
}

// ── user data ──────────────────────────────────────────────────────────────────

var nuUsers = []uDef{
	{"Айгерім Бекова", "aiguerim@nu.edu.kz", entities.RoleAdmin},
	{"Бауыржан Сейітов", "baurzhan@nu.edu.kz", entities.RoleTeacher},
	{"Дамир Жаксыбеков", "damir@nu.edu.kz", entities.RoleTeacher},
	{"Гүлнар Мұратова", "gulnar@nu.edu.kz", entities.RoleTeacher},
	{"Ерлан Абдуллаев", "erlan@nu.edu.kz", entities.RoleTeacher},
	{"Зарина Нұрмаханова", "zarina@nu.edu.kz", entities.RoleTeacher},
	{"Санжар Мұқанов", "sanzhar@nu.edu.kz", entities.RoleStudent},
	{"Жанар Нұрланова", "zhanar@nu.edu.kz", entities.RoleStudent},
	{"Асхат Бердібеков", "askhat@nu.edu.kz", entities.RoleStudent},
	{"Айнұр Тасанова", "ainur@nu.edu.kz", entities.RoleStudent},
	{"Дидар Сейітов", "didar@nu.edu.kz", entities.RoleStudent},
	{"Ұлбосын Тоқтарова", "ulbosyn@nu.edu.kz", entities.RoleStudent},
	{"Назым Абенова", "nazym@nu.edu.kz", entities.RoleStudent},
	{"Серік Жақсыбаев", "serik@nu.edu.kz", entities.RoleStudent},
	{"Ақбота Мырзабекова", "akbota@nu.edu.kz", entities.RoleStudent},
	{"Данияр Есенов", "daniyar@nu.edu.kz", entities.RoleStudent},
	{"Гүлбану Алибекова", "gulbanu@nu.edu.kz", entities.RoleStudent},
	{"Нұрсат Ермеков", "nursat@nu.edu.kz", entities.RoleStudent},
	{"Зере Ахметова", "zere@nu.edu.kz", entities.RoleStudent},
	{"Арман Досжанов", "arman@nu.edu.kz", entities.RoleStudent},
	{"Мадина Қасымова", "madina@nu.edu.kz", entities.RoleStudent},
	{"Темірлан Ізтелеуов", "temirlan@nu.edu.kz", entities.RoleStudent},
	{"Аружан Сейтқали", "aruzhan@nu.edu.kz", entities.RoleStudent},
	{"Олжас Бейсенбаев", "olzhas@nu.edu.kz", entities.RoleStudent},
	{"Камила Жанабекова", "kamila@nu.edu.kz", entities.RoleStudent},
	{"Ербол Нұрланов", "erbol@nu.edu.kz", entities.RoleStudent},
	{"Сандугаш Омарова", "sandugash@nu.edu.kz", entities.RoleStudent},
	{"Азамат Байжанов", "azamat@nu.edu.kz", entities.RoleStudent},
	{"Лейла Абдрахманова", "leila@nu.edu.kz", entities.RoleStudent},
	{"Қайрат Мұхамеджанов", "kairat@nu.edu.kz", entities.RoleStudent},
	{"Айгерім Сейтжанова", "aigerim.s@nu.edu.kz", entities.RoleStudent},
}

var kbtuUsers = []uDef{
	{"Маржан Сейтқали", "marzhan@kbtu.edu.kz", entities.RoleAdmin},
	{"Нұрлан Дүйсебаев", "nurlan@kbtu.edu.kz", entities.RoleTeacher},
	{"Айбек Жалмағанбет", "aibek@kbtu.edu.kz", entities.RoleTeacher},
	{"Салтанат Бейсенбаева", "saltanat@kbtu.edu.kz", entities.RoleTeacher},
	{"Рустем Қожабеков", "rustem@kbtu.edu.kz", entities.RoleTeacher},
	{"Жулдыз Мәдиева", "zhuldyz@kbtu.edu.kz", entities.RoleTeacher},
	{"Нұрдаулет Ахмедов", "nurdaulet@kbtu.edu.kz", entities.RoleStudent},
	{"Сабина Джаксыбекова", "sabina@kbtu.edu.kz", entities.RoleStudent},
	{"Дәурен Исабеков", "dauren@kbtu.edu.kz", entities.RoleStudent},
	{"Аида Мирасбекова", "aida@kbtu.edu.kz", entities.RoleStudent},
	{"Берік Сапаров", "berik@kbtu.edu.kz", entities.RoleStudent},
	{"Жансая Қалиева", "zhansaya@kbtu.edu.kz", entities.RoleStudent},
	{"Нурас Бекмуханов", "nuras@kbtu.edu.kz", entities.RoleStudent},
	{"Меруерт Тілеуберді", "meruert@kbtu.edu.kz", entities.RoleStudent},
	{"Ілияс Жетписбаев", "iliyas@kbtu.edu.kz", entities.RoleStudent},
	{"Дина Ахметбекова", "dina@kbtu.edu.kz", entities.RoleStudent},
	{"Болат Ердесов", "bolat@kbtu.edu.kz", entities.RoleStudent},
	{"Жазира Мамытбекова", "zhazira@kbtu.edu.kz", entities.RoleStudent},
	{"Нұрлан Темірбеков", "nurlan.t@kbtu.edu.kz", entities.RoleStudent},
	{"Алина Сейткали", "alina@kbtu.edu.kz", entities.RoleStudent},
	{"Думан Тоқтасынов", "duman@kbtu.edu.kz", entities.RoleStudent},
	{"Бибігүл Жақсылықова", "bibigul@kbtu.edu.kz", entities.RoleStudent},
	{"Талғат Мұсаев", "talgat@kbtu.edu.kz", entities.RoleStudent},
	{"Молдір Нурмаханова", "moldir@kbtu.edu.kz", entities.RoleStudent},
	{"Жандос Байбосынов", "zhandos@kbtu.edu.kz", entities.RoleStudent},
	{"Шынар Алиасқарова", "shynar@kbtu.edu.kz", entities.RoleStudent},
	{"Эрик Садуақасов", "erik@kbtu.edu.kz", entities.RoleStudent},
	{"Гүлмира Асанова", "gulmira@kbtu.edu.kz", entities.RoleStudent},
	{"Мейрам Сейтберген", "meiram@kbtu.edu.kz", entities.RoleStudent},
	{"Раушан Болатбекова", "raushan@kbtu.edu.kz", entities.RoleStudent},
	{"Сырым Аширбеков", "syrym@kbtu.edu.kz", entities.RoleStudent},
}

// ── main ───────────────────────────────────────────────────────────────────────

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(slog.LevelInfo, "text")

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := database.RunMigrations(db); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	wipe(db, log)

	d := buildDeps(db)
	rng := rand.New(rand.NewSource(42))
	ctx := context.Background()

	seedOrg(ctx, d, "Назарбаев Университеті", "nu", nuUsers, rng, log)
	seedOrg(ctx, d, "Қазақ-Британ Техникалық Университеті", "kbtu", kbtuUsers, rng, log)

	fmt.Println()
	fmt.Println("✓ Seed complete!")
	fmt.Println()
	printCredentials()
}

// ── wipe ───────────────────────────────────────────────────────────────────────

func wipe(db *gorm.DB, log *logger.Logger) {
	log.Info("wiping database...")
	err := db.Exec(`TRUNCATE TABLE
		attendance_records,
		lesson_progress,
		announcements,
		group_members,
		group_schedules,
		groups,
		extra_attempt_grants,
		exam_attempts,
		exam_answers,
		exam_questions,
		exams,
		assignment_submissions,
		assignments,
		quiz_attempts,
		quiz_answers,
		quiz_questions,
		quizzes,
		file_attachments,
		lessons,
		modules,
		course_teachers,
		courses,
		memberships,
		users,
		organizations
		RESTART IDENTITY CASCADE`).Error
	if err != nil {
		log.Error("wipe failed", "error", err)
		os.Exit(1)
	}
	log.Info("database wiped")
}

// ── buildDeps ──────────────────────────────────────────────────────────────────

func buildDeps(db *gorm.DB) seedDeps {
	return seedDeps{
		userRepo:     infraRepos.NewGORMUserRepository(db),
		orgRepo:      infraRepos.NewGORMOrganizationRepository(db),
		memberRepo:   infraRepos.NewGORMMembershipRepository(db),
		courseRepo:   infraRepos.NewGORMCourseRepository(db),
		ctRepo:       infraRepos.NewGORMCourseTeacherRepository(db),
		moduleRepo:   infraRepos.NewGORMModuleRepository(db),
		lessonRepo:   infraRepos.NewGORMLessonRepository(db),
		quizRepo:     infraRepos.NewGORMQuizRepository(db),
		qaRepo:       infraRepos.NewGORMQuizAttemptRepository(db),
		assignRepo:   infraRepos.NewGORMAssignmentRepository(db),
		examRepo:     infraRepos.NewGORMExamRepository(db),
		eaRepo:       infraRepos.NewGORMExamAttemptRepository(db),
		grantRepo:    infraRepos.NewGORMExtraAttemptGrantRepository(db),
		groupRepo:    infraRepos.NewGORMGroupRepository(db),
		announceRepo: infraRepos.NewGORMAnnouncementRepository(db),
		progressRepo: infraRepos.NewGORMProgressRepository(db),
	}
}

// ── seedOrg ────────────────────────────────────────────────────────────────────

func seedOrg(ctx context.Context, d seedDeps, orgName, orgSlug string, userDefs []uDef, rng *rand.Rand, log *logger.Logger) {
	log.Info("seeding org", "slug", orgSlug)
	now := time.Now()

	org := &entities.Organization{
		ID:        uuid.New(),
		Name:      orgName,
		Slug:      orgSlug,
		CreatedAt: daysAgo(90),
		UpdatedAt: now,
	}
	if err := d.orgRepo.Create(ctx, org); err != nil {
		log.Error("failed to create org", "slug", orgSlug, "error", err)
		os.Exit(1)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	pwHash := string(hash)

	var admin *entities.User
	var teachers []*entities.User
	var students []*entities.User

	for _, ud := range userDefs {
		u := &entities.User{
			ID:           uuid.New(),
			Email:        ud.email,
			PasswordHash: pwHash,
			Name:         ud.name,
			CreatedAt:    daysAgo(80),
			UpdatedAt:    now,
		}
		if err := d.userRepo.Create(ctx, u); err != nil {
			log.Error("failed to create user", "email", ud.email, "error", err)
			continue
		}
		m := &entities.Membership{
			ID:        uuid.New(),
			UserID:    u.ID,
			OrgID:     org.ID,
			Role:      ud.role,
			CreatedAt: daysAgo(80),
			UpdatedAt: now,
		}
		if err := d.memberRepo.Create(ctx, m); err != nil {
			log.Error("failed to create membership", "email", ud.email, "error", err)
		}
		switch ud.role {
		case entities.RoleAdmin:
			admin = u
		case entities.RoleTeacher:
			teachers = append(teachers, u)
		case entities.RoleStudent:
			students = append(students, u)
		}
	}

	if admin == nil || len(teachers) == 0 || len(students) == 0 {
		log.Error("missing users", "slug", orgSlug)
		os.Exit(1)
	}

	log.Info("users created", "slug", orgSlug, "teachers", len(teachers), "students", len(students))

	// ── courses ────────────────────────────────────────────────────────────────
	for ci, cd := range courseTemplates {
		teacher := teachers[ci%len(teachers)]
		courseCreatedAt := daysAgo(60 + rng.Intn(20))

		course := &entities.Course{
			ID:          uuid.New(),
			OrgID:       org.ID,
			Title:       cd.title,
			Description: cd.desc,
			Status:      cd.status,
			CreatedBy:   admin.ID,
			CreatedAt:   courseCreatedAt,
			UpdatedAt:   courseCreatedAt,
		}
		if err := d.courseRepo.Create(ctx, course); err != nil {
			log.Error("failed to create course", "title", cd.title, "error", err)
			continue
		}

		ct := &entities.CourseTeacher{
			ID:         uuid.New(),
			CourseID:   course.ID,
			TeacherID:  teacher.ID,
			OrgID:      org.ID,
			AssignedAt: courseCreatedAt,
		}
		if err := d.ctRepo.Add(ctx, ct); err != nil {
			log.Error("failed to assign teacher", "course", cd.title, "error", err)
		}

		// ── modules & lessons ─────────────────────────────────────────────────
		type createdMod struct {
			mod     *entities.Module
			lessons [4]*entities.Lesson
		}
		var allMods [3]createdMod

		for mi, md := range cd.mods {
			mod := &entities.Module{
				ID:        uuid.New(),
				CourseID:  course.ID,
				OrgID:     org.ID,
				Title:     md.title,
				Position:  mi + 1,
				CreatedAt: daysAgo(50 + rng.Intn(10)),
				UpdatedAt: now,
			}
			if err := d.moduleRepo.Create(ctx, mod); err != nil {
				log.Error("failed to create module", "error", err)
				continue
			}
			allMods[mi].mod = mod

			for li, ld := range md.lessons {
				lesson := &entities.Lesson{
					ID:        uuid.New(),
					ModuleID:  mod.ID,
					OrgID:     org.ID,
					Title:     ld.title,
					Content:   ld.content,
					Type:      ld.typ,
					VideoURL:  ld.video,
					FileURL:   ld.file,
					LinkURL:   ld.link,
					Position:  li + 1,
					CreatedAt: daysAgo(45 + rng.Intn(10)),
					UpdatedAt: now,
				}
				if err := d.lessonRepo.Create(ctx, lesson); err != nil {
					log.Error("failed to create lesson", "error", err)
					continue
				}
				allMods[mi].lessons[li] = lesson
			}
		}

		if cd.status != entities.StatusPublished {
			continue
		}

		// ── quizzes ───────────────────────────────────────────────────────────
		var quizzes [3]*entities.Quiz
		for mi, md := range cd.mods {
			firstLesson := allMods[mi].lessons[0]
			if firstLesson == nil || len(md.lessons[0].quiz) == 0 {
				continue
			}
			quiz := buildQuiz(firstLesson.ID, org.ID, firstLesson.Title, md.lessons[0].quiz, now)
			if err := d.quizRepo.Create(ctx, quiz); err != nil {
				log.Error("failed to create quiz", "error", err)
				continue
			}
			quizzes[mi] = quiz
		}

		// ── assignments ───────────────────────────────────────────────────────
		var assignments [2]*entities.Assignment
		for ai, ad := range cd.assigns {
			mod := allMods[ai].mod
			if mod == nil {
				continue
			}
			dueDate := daysFromNow(ad.days)
			a := &entities.Assignment{
				ID:                  uuid.New(),
				OrgID:               org.ID,
				CourseID:            course.ID,
				ModuleID:            mod.ID,
				Title:               ad.title,
				Description:         ad.desc,
				MaxPoints:           ad.pts,
				DueDate:             &dueDate,
				AllowLateSubmission: true,
				Position:            ai + 1,
				CreatedAt:           daysAgo(40 + rng.Intn(10)),
				UpdatedAt:           now,
			}
			if err := d.assignRepo.Create(ctx, a); err != nil {
				log.Error("failed to create assignment", "error", err)
				continue
			}
			assignments[ai] = a
		}

		// ── exams ─────────────────────────────────────────────────────────────
		examDueDays := [2]int{20, 45}
		var exams [2]*entities.Exam
		for ei, ed := range cd.exams {
			dueDate := daysFromNow(examDueDays[ei])
			exam := buildExam(course.ID, org.ID, ed, dueDate, now)
			if err := d.examRepo.Create(ctx, exam); err != nil {
				log.Error("failed to create exam", "error", err)
				continue
			}
			exams[ei] = exam
		}

		// ── groups ────────────────────────────────────────────────────────────
		groupDays := [2]int{1, 4}
		groupTimes := [2][2]string{{"09:00", "10:30"}, {"14:00", "15:30"}}
		for gi := 0; gi < 3; gi++ {
			cid := course.ID
			tid := teacher.ID
			g := &entities.Group{
				ID:        uuid.New(),
				OrgID:     org.ID,
				CourseID:  &cid,
				TeacherID: &tid,
				Name:      fmt.Sprintf("%s – %d-топ", course.Title, gi+1),
				CreatedAt: daysAgo(55),
				UpdatedAt: now,
			}
			if err := d.groupRepo.CreateGroup(ctx, g); err != nil {
				log.Error("failed to create group", "error", err)
				continue
			}
			for si, t := range groupTimes {
				sched := &entities.GroupSchedule{
					ID:        uuid.New(),
					GroupID:   g.ID,
					DayOfWeek: groupDays[si],
					StartTime: t[0],
					EndTime:   t[1],
					Location:  fmt.Sprintf("Аудитория %d%02d", gi+1, (gi+1)*10+si+1),
					CreatedAt: daysAgo(55),
				}
				if err := d.groupRepo.AddSchedule(ctx, sched); err != nil {
					log.Error("failed to add schedule", "error", err)
				}
			}
			for si, student := range students {
				if si%3 != gi {
					continue
				}
				member := &entities.GroupMember{
					ID:        uuid.New(),
					GroupID:   g.ID,
					StudentID: student.ID,
					OrgID:     org.ID,
					JoinedAt:  daysAgo(rng.Intn(50) + 5),
				}
				if err := d.groupRepo.AddMember(ctx, member); err != nil {
					log.Error("failed to add group member", "error", err)
				}
			}
		}

		// ── announcements ─────────────────────────────────────────────────────
		for _, ann := range cd.anns {
			a := &entities.Announcement{
				ID:        uuid.New(),
				CourseID:  course.ID,
				OrgID:     org.ID,
				AuthorID:  teacher.ID,
				Title:     ann.title,
				Content:   ann.body,
				CreatedAt: daysAgo(rng.Intn(45) + 1),
				UpdatedAt: now,
			}
			if err := d.announceRepo.Create(ctx, a); err != nil {
				log.Error("failed to create announcement", "error", err)
			}
		}

		// ── quiz attempts ─────────────────────────────────────────────────────
		for mi, quiz := range quizzes {
			if quiz == nil {
				continue
			}
			for si, student := range students {
				if rng.Float64() > 0.80 {
					continue
				}
				scorePct := randScorePct(rng, si)
				answers, correctCount := buildQuizAttemptAnswers(rng, quiz, scorePct)
				scorePerQ := quiz.MaxPoints / len(quiz.Questions)
				score := correctCount * scorePerQ
				submittedAt := daysAgo(rng.Intn(30) + 1)
				attempt := &entities.QuizAttempt{
					ID:          uuid.New(),
					QuizID:      quiz.ID,
					StudentID:   student.ID,
					OrgID:       org.ID,
					Score:       score,
					MaxScore:    quiz.MaxPoints,
					Answers:     answers,
					SubmittedAt: submittedAt,
					CreatedAt:   submittedAt,
					UpdatedAt:   submittedAt,
				}
				_ = mi
				if err := d.qaRepo.Create(ctx, attempt); err != nil {
					log.Error("failed to create quiz attempt", "error", err)
				}
			}
		}

		// ── assignment submissions ─────────────────────────────────────────────
		submissionTexts := []string{
			"Тапсырманы орындадым. Барлық талаптар сақталды. Код GitHub-та.",
			"Міне менің шешімім. Тестілер өтті, қателер жоқ.",
			"Тапсырма дайын. Қосымша функциялар да іске асырылды.",
			"Барлық талаптарды орындадым. Сілтеме: github.com/example/repo",
			"Жұмыс аяқталды. PDF форматында есеп қоса берілді.",
		}
		feedbacks := []string{
			"Жақсы жұмыс! Код таза жазылған.",
			"Өте жақсы! Барлық талаптар сақталды.",
			"Жақсы, бірақ қателерді өңдеу жақсартуды қажет.",
			"Орташа деңгей. Кейбір функциялар жетіспейді.",
			"Мықты жұмыс! Қосымша функциялар бонус берді.",
		}
		for _, assignment := range assignments {
			if assignment == nil {
				continue
			}
			for si, student := range students {
				if rng.Float64() > 0.70 {
					continue
				}
				submittedAt := daysAgo(rng.Intn(14) + 1)
				sub := &entities.AssignmentSubmission{
					ID:           uuid.New(),
					AssignmentID: assignment.ID,
					StudentID:    student.ID,
					OrgID:        org.ID,
					TextContent:  submissionTexts[rng.Intn(len(submissionTexts))],
					SubmittedAt:  submittedAt,
					CreatedAt:    submittedAt,
					UpdatedAt:    submittedAt,
				}
				if err := d.assignRepo.CreateSubmission(ctx, sub); err != nil {
					log.Error("failed to create submission", "error", err)
					continue
				}
				if rng.Float64() < 0.55 {
					gradedAt := submittedAt.Add(time.Duration(rng.Intn(72)+24) * time.Hour)
					scorePct := randScorePct(rng, si)
					score := int(scorePct * float64(assignment.MaxPoints))
					sub.Score = &score
					sub.Feedback = feedbacks[rng.Intn(len(feedbacks))]
					sub.GradedBy = &teacher.ID
					sub.GradedAt = &gradedAt
					if err := d.assignRepo.UpdateSubmission(ctx, sub); err != nil {
						log.Error("failed to grade submission", "error", err)
					}
				}
			}
		}

		// ── exam attempts ─────────────────────────────────────────────────────
		for _, exam := range exams {
			if exam == nil {
				continue
			}
			var lowestScore int = 100
			var lowestStudent *entities.User
			for si, student := range students {
				if rng.Float64() > 0.60 {
					continue
				}
				scorePct := randScorePct(rng, si)
				r := rng.Float64()

				var status string
				var submittedAt *time.Time
				var mcqScore, totalScore *int
				var gradedBy *uuid.UUID
				var gradedAt *time.Time
				var mcqAnswers []entities.ExamMCQAnswer

				startedAt := daysAgo(rng.Intn(25) + 1)
				expiresAt := startedAt.Add(time.Duration(exam.DurationMinutes) * time.Minute)

				switch {
				case r < 0.10:
					status = "in_progress"
				case r < 0.35:
					status = "submitted"
					t := startedAt.Add(time.Duration(exam.DurationMinutes/2+rng.Intn(exam.DurationMinutes/2)) * time.Minute)
					submittedAt = &t
					mcqAnswers, _ = buildExamMCQAnswers(rng, exam, scorePct)
				default:
					status = "submitted"
					t := startedAt.Add(time.Duration(exam.DurationMinutes/2+rng.Intn(exam.DurationMinutes/2)) * time.Minute)
					submittedAt = &t
					var correctCount int
					mcqAnswers, correctCount = buildExamMCQAnswers(rng, exam, scorePct)
					sc := 0
					if len(exam.Questions) > 0 {
						sc = int(float64(correctCount) / float64(len(exam.Questions)) * float64(exam.MCQPoints))
					}
					mcqScore = ptrInt(sc)
					totalScore = ptrInt(sc)
					gt := t.Add(time.Duration(rng.Intn(48)+24) * time.Hour)
					gradedAt = &gt
					gradedBy = &teacher.ID

					if sc < lowestScore {
						lowestScore = sc
						lowestStudent = student
					}
				}

				attempt := &entities.ExamAttempt{
					ID:          uuid.New(),
					ExamID:      exam.ID,
					StudentID:   student.ID,
					OrgID:       org.ID,
					Status:      status,
					StartedAt:   startedAt,
					ExpiresAt:   expiresAt,
					SubmittedAt: submittedAt,
					MCQAnswers:  mcqAnswers,
					MCQScore:    mcqScore,
					MCQMaxScore: exam.MCQPoints,
					TotalScore:  totalScore,
					GradedBy:    gradedBy,
					GradedAt:    gradedAt,
					CreatedAt:   startedAt,
					UpdatedAt:   startedAt,
				}
				if err := d.eaRepo.Create(ctx, attempt); err != nil {
					log.Error("failed to create exam attempt", "error", err)
				}
			}

			// grant extra attempt to lowest-scoring student (if below 60%)
			threshold := int(0.60 * float64(exam.MCQPoints))
			if lowestStudent != nil && lowestScore < threshold {
				grant := &entities.ExtraAttemptGrant{
					ID:         uuid.New(),
					ExamID:     exam.ID,
					StudentID:  lowestStudent.ID,
					OrgID:      org.ID,
					GrantedBy:  teacher.ID,
					ExtraCount: 1,
					CreatedAt:  daysAgo(rng.Intn(10) + 1),
				}
				if err := d.grantRepo.Create(ctx, grant); err != nil {
					log.Error("failed to create extra grant", "error", err)
				}
			}
		}

		// ── lesson progress ───────────────────────────────────────────────────
		for mi := range cd.mods {
			for li := 0; li < 4; li++ {
				lesson := allMods[mi].lessons[li]
				if lesson == nil {
					continue
				}
				for si, student := range students {
					threshold := 0.85 - float64(si)*0.024
					if rng.Float64() >= threshold {
						continue
					}
					score := 55.0 + rng.Float64()*45.0 - float64(si)*0.5
					if score < 40.0 {
						score = 40.0
					}
					completedAt := daysAgo(rng.Intn(40) + 1)
					p := &entities.LessonProgress{
						ID:          uuid.New(),
						UserID:      student.ID,
						LessonID:    lesson.ID,
						OrgID:       org.ID,
						CompletedAt: &completedAt,
						Score:       &score,
						CreatedAt:   completedAt,
						UpdatedAt:   completedAt,
					}
					if err := d.progressRepo.Create(ctx, p); err != nil {
						log.Error("failed to create progress", "error", err)
					}
				}
			}
		}

		log.Info("course seeded", "title", cd.title, "org", orgSlug)
	}

	log.Info("org seeded", "slug", orgSlug)
}

// ── helpers ────────────────────────────────────────────────────────────────────

func buildQuiz(lessonID, orgID uuid.UUID, lessonTitle string, qdefs []qDef, now time.Time) *entities.Quiz {
	questions := make([]entities.QuizQuestion, len(qdefs))
	for i, qd := range qdefs {
		qID := uuid.New()
		answers := []entities.QuizAnswer{
			{ID: uuid.New(), QuestionID: qID, Answer: qd.correct, IsCorrect: true, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: qID, Answer: qd.wrong[0], IsCorrect: false, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: qID, Answer: qd.wrong[1], IsCorrect: false, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: qID, Answer: qd.wrong[2], IsCorrect: false, CreatedAt: now, UpdatedAt: now},
		}
		questions[i] = entities.QuizQuestion{
			ID:        qID,
			Question:  qd.q,
			Position:  i + 1,
			Answers:   answers,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	maxPts := len(qdefs) * 5
	return &entities.Quiz{
		ID:        uuid.New(),
		LessonID:  lessonID,
		OrgID:     orgID,
		Title:     lessonTitle + " – Тексеру сұрақтары",
		MaxPoints: maxPts,
		Questions: questions,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildExam(courseID, orgID uuid.UUID, ed eDef, dueDate time.Time, now time.Time) *entities.Exam {
	questions := make([]entities.ExamQuestion, len(ed.qs))
	for i, qd := range ed.qs {
		qID := uuid.New()
		answers := []entities.ExamAnswer{
			{ID: uuid.New(), QuestionID: qID, Answer: qd.correct, IsCorrect: true, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: qID, Answer: qd.wrong[0], IsCorrect: false, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: qID, Answer: qd.wrong[1], IsCorrect: false, CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), QuestionID: qID, Answer: qd.wrong[2], IsCorrect: false, CreatedAt: now, UpdatedAt: now},
		}
		questions[i] = entities.ExamQuestion{
			ID:        qID,
			Question:  qd.q,
			Position:  i + 1,
			Answers:   answers,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return &entities.Exam{
		ID:              uuid.New(),
		CourseID:        courseID,
		OrgID:           orgID,
		Title:           ed.title,
		Description:     ed.desc,
		DurationMinutes: ed.mins,
		MaxAttempts:     ed.maxAttempts,
		MCQEnabled:      true,
		MCQPoints:       ed.pts,
		DueDate:         &dueDate,
		Questions:       questions,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func buildQuizAttemptAnswers(rng *rand.Rand, quiz *entities.Quiz, scorePct float64) ([]entities.QuizAttemptAnswer, int) {
	correctCount := 0
	answers := make([]entities.QuizAttemptAnswer, len(quiz.Questions))
	for i, q := range quiz.Questions {
		var correctID, wrongID string
		for _, a := range q.Answers {
			if a.IsCorrect {
				correctID = a.ID.String()
			} else if wrongID == "" {
				wrongID = a.ID.String()
			}
		}
		chosenID := wrongID
		if rng.Float64() < scorePct {
			chosenID = correctID
			correctCount++
		}
		answers[i] = entities.QuizAttemptAnswer{
			QuestionID: q.ID.String(),
			AnswerID:   chosenID,
		}
	}
	return answers, correctCount
}

func buildExamMCQAnswers(rng *rand.Rand, exam *entities.Exam, scorePct float64) ([]entities.ExamMCQAnswer, int) {
	correctCount := 0
	answers := make([]entities.ExamMCQAnswer, len(exam.Questions))
	for i, q := range exam.Questions {
		var correctID, wrongID string
		for _, a := range q.Answers {
			if a.IsCorrect {
				correctID = a.ID.String()
			} else if wrongID == "" {
				wrongID = a.ID.String()
			}
		}
		chosenID := wrongID
		if rng.Float64() < scorePct {
			chosenID = correctID
			correctCount++
		}
		answers[i] = entities.ExamMCQAnswer{
			QuestionID: q.ID.String(),
			AnswerID:   chosenID,
		}
	}
	return answers, correctCount
}

func randScorePct(rng *rand.Rand, studentIdx int) float64 {
	base := 1.0 - float64(studentIdx)*0.012
	r := rng.Float64()
	switch {
	case r < 0.25:
		return clamp(base, 0.3, 1.0)
	case r < 0.75:
		return clamp(base-0.15+rng.Float64()*0.25, 0.3, 1.0)
	case r < 0.95:
		return clamp(base-0.30+rng.Float64()*0.20, 0.3, 1.0)
	default:
		return clamp(base-0.50+rng.Float64()*0.20, 0.3, 0.60)
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func ptrInt(n int) *int { return &n }

func daysAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * 24 * time.Hour)
}

func daysFromNow(n int) time.Time {
	return time.Now().Add(time.Duration(n) * 24 * time.Hour)
}

func printCredentials() {
	fmt.Println("  ╔══════════════════════════════════════════════════════════╗")
	fmt.Println("  ║               Тіркелу деректері / Credentials           ║")
	fmt.Println("  ╠══════════════════════════════════════════════════════════╣")
	fmt.Println("  ║ Org: nu  (Назарбаев Университеті)                       ║")
	fmt.Println("  ║  admin    aiguerim@nu.edu.kz        password123         ║")
	fmt.Println("  ║  teacher  baurzhan@nu.edu.kz         password123         ║")
	fmt.Println("  ║  teacher  damir@nu.edu.kz            password123         ║")
	fmt.Println("  ║  student  sanzhar@nu.edu.kz          password123         ║")
	fmt.Println("  ║  student  zhanar@nu.edu.kz           password123         ║")
	fmt.Println("  ╠══════════════════════════════════════════════════════════╣")
	fmt.Println("  ║ Org: kbtu (Қазақ-Британ Техникалық Университеті)        ║")
	fmt.Println("  ║  admin    marzhan@kbtu.edu.kz       password123         ║")
	fmt.Println("  ║  teacher  nurlan@kbtu.edu.kz         password123         ║")
	fmt.Println("  ║  student  nurdaulet@kbtu.edu.kz      password123         ║")
	fmt.Println("  ║  student  sabina@kbtu.edu.kz         password123         ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  X-Org-Slug header: nu  немесе  kbtu")
}
