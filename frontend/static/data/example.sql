PGDMP  9                    }            user_management    17.4    17.4 }    m           0    0    ENCODING    ENCODING        SET client_encoding = 'UTF8';
                           false            n           0    0 
   STDSTRINGS 
   STDSTRINGS     (   SET standard_conforming_strings = 'on';
                           false            o           0    0 
   SEARCHPATH 
   SEARCHPATH     8   SELECT pg_catalog.set_config('search_path', '', false);
                           false            p           1262    16388    user_management    DATABASE     u   CREATE DATABASE user_management WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'ru-RU';
    DROP DATABASE user_management;
                     postgres    false            �            1259    17114 
   categories    TABLE     �   CREATE TABLE public.categories (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
    DROP TABLE public.categories;
       public         heap r       postgres    false            �            1259    17113    categories_id_seq    SEQUENCE     �   CREATE SEQUENCE public.categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 (   DROP SEQUENCE public.categories_id_seq;
       public               postgres    false    232            q           0    0    categories_id_seq    SEQUENCE OWNED BY     G   ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;
          public               postgres    false    231            �            1259    17125    dishes    TABLE     �  CREATE TABLE public.dishes (
    id integer NOT NULL,
    category_id integer,
    name character varying(100) NOT NULL,
    price numeric(10,2) NOT NULL,
    image_url text DEFAULT ''::text NOT NULL,
    is_available boolean DEFAULT true,
    preparation_time integer,
    calories integer,
    allergens text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
    DROP TABLE public.dishes;
       public         heap r       postgres    false            �            1259    17124    dishes_id_seq    SEQUENCE     �   CREATE SEQUENCE public.dishes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 $   DROP SEQUENCE public.dishes_id_seq;
       public               postgres    false    234            r           0    0    dishes_id_seq    SEQUENCE OWNED BY     ?   ALTER SEQUENCE public.dishes_id_seq OWNED BY public.dishes.id;
          public               postgres    false    233            �            1259    16928 	   inventory    TABLE     �  CREATE TABLE public.inventory (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    category character varying(100) NOT NULL,
    quantity numeric(10,2) NOT NULL,
    unit character varying(10) NOT NULL,
    min_quantity numeric(10,2) NOT NULL,
    branch character varying(100),
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
    DROP TABLE public.inventory;
       public         heap r       postgres    false            �            1259    16927    inventory_id_seq    SEQUENCE     �   CREATE SEQUENCE public.inventory_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 '   DROP SEQUENCE public.inventory_id_seq;
       public               postgres    false    224            s           0    0    inventory_id_seq    SEQUENCE OWNED BY     E   ALTER SEQUENCE public.inventory_id_seq OWNED BY public.inventory.id;
          public               postgres    false    223            �            1259    17142    menu_item_options    TABLE     �   CREATE TABLE public.menu_item_options (
    id integer NOT NULL,
    dish_id integer,
    name character varying(100) NOT NULL,
    price_modifier numeric(10,2) DEFAULT 0,
    is_available boolean DEFAULT true
);
 %   DROP TABLE public.menu_item_options;
       public         heap r       postgres    false            �            1259    17141    menu_item_options_id_seq    SEQUENCE     �   CREATE SEQUENCE public.menu_item_options_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 /   DROP SEQUENCE public.menu_item_options_id_seq;
       public               postgres    false    236            t           0    0    menu_item_options_id_seq    SEQUENCE OWNED BY     U   ALTER SEQUENCE public.menu_item_options_id_seq OWNED BY public.menu_item_options.id;
          public               postgres    false    235            �            1259    17177    order_item_options    TABLE     v   CREATE TABLE public.order_item_options (
    id integer NOT NULL,
    order_item_id integer,
    option_id integer
);
 &   DROP TABLE public.order_item_options;
       public         heap r       postgres    false            �            1259    17176    order_item_options_id_seq    SEQUENCE     �   CREATE SEQUENCE public.order_item_options_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 0   DROP SEQUENCE public.order_item_options_id_seq;
       public               postgres    false    240            u           0    0    order_item_options_id_seq    SEQUENCE OWNED BY     W   ALTER SEQUENCE public.order_item_options_id_seq OWNED BY public.order_item_options.id;
          public               postgres    false    239            �            1259    17156    order_items    TABLE     -  CREATE TABLE public.order_items (
    id integer NOT NULL,
    order_id integer,
    dish_id integer,
    quantity integer NOT NULL,
    price numeric(10,2) NOT NULL,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
    DROP TABLE public.order_items;
       public         heap r       postgres    false            �            1259    17155    order_items_id_seq    SEQUENCE     �   CREATE SEQUENCE public.order_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 )   DROP SEQUENCE public.order_items_id_seq;
       public               postgres    false    238            v           0    0    order_items_id_seq    SEQUENCE OWNED BY     I   ALTER SEQUENCE public.order_items_id_seq OWNED BY public.order_items.id;
          public               postgres    false    237            �            1259    16888    orders    TABLE     �  CREATE TABLE public.orders (
    id integer NOT NULL,
    table_id integer NOT NULL,
    waiter_id integer NOT NULL,
    status character varying(20) DEFAULT 'new'::character varying NOT NULL,
    total_amount integer NOT NULL,
    comment text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    completed_at timestamp without time zone,
    cancelled_at timestamp without time zone
);
    DROP TABLE public.orders;
       public         heap r       postgres    false            �            1259    16887    orders_id_seq    SEQUENCE     �   CREATE SEQUENCE public.orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 $   DROP SEQUENCE public.orders_id_seq;
       public               postgres    false    222            w           0    0    orders_id_seq    SEQUENCE OWNED BY     ?   ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;
          public               postgres    false    221            �            1259    16966    request_items    TABLE     �   CREATE TABLE public.request_items (
    id integer NOT NULL,
    request_id integer,
    inventory_id integer,
    quantity numeric(10,2)
);
 !   DROP TABLE public.request_items;
       public         heap r       postgres    false            �            1259    16965    request_items_id_seq    SEQUENCE     �   CREATE SEQUENCE public.request_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 +   DROP SEQUENCE public.request_items_id_seq;
       public               postgres    false    230            x           0    0    request_items_id_seq    SEQUENCE OWNED BY     M   ALTER SEQUENCE public.request_items_id_seq OWNED BY public.request_items.id;
          public               postgres    false    229            �            1259    16949    requests    TABLE     �  CREATE TABLE public.requests (
    id integer NOT NULL,
    branch character varying(100) NOT NULL,
    supplier_id integer,
    items text[],
    priority character varying(20) DEFAULT 'normal'::character varying,
    comment text,
    status character varying(20) DEFAULT 'pending'::character varying,
    created_at timestamp without time zone DEFAULT now(),
    completed_at timestamp without time zone
);
    DROP TABLE public.requests;
       public         heap r       postgres    false            �            1259    16948    requests_id_seq    SEQUENCE     �   CREATE SEQUENCE public.requests_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 &   DROP SEQUENCE public.requests_id_seq;
       public               postgres    false    228            y           0    0    requests_id_seq    SEQUENCE OWNED BY     C   ALTER SEQUENCE public.requests_id_seq OWNED BY public.requests.id;
          public               postgres    false    227            �            1259    17401    shift_employees    TABLE     �   CREATE TABLE public.shift_employees (
    id integer NOT NULL,
    shift_id integer NOT NULL,
    employee_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);
 #   DROP TABLE public.shift_employees;
       public         heap r       postgres    false            �            1259    17400    shift_employees_id_seq    SEQUENCE     �   CREATE SEQUENCE public.shift_employees_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 -   DROP SEQUENCE public.shift_employees_id_seq;
       public               postgres    false    244            z           0    0    shift_employees_id_seq    SEQUENCE OWNED BY     Q   ALTER SEQUENCE public.shift_employees_id_seq OWNED BY public.shift_employees.id;
          public               postgres    false    243            �            1259    17385    shifts    TABLE     S  CREATE TABLE public.shifts (
    id integer NOT NULL,
    date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    manager_id integer NOT NULL,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
    DROP TABLE public.shifts;
       public         heap r       postgres    false            �            1259    17384    shifts_id_seq    SEQUENCE     �   CREATE SEQUENCE public.shifts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 $   DROP SEQUENCE public.shifts_id_seq;
       public               postgres    false    242            {           0    0    shifts_id_seq    SEQUENCE OWNED BY     ?   ALTER SEQUENCE public.shifts_id_seq OWNED BY public.shifts.id;
          public               postgres    false    241            �            1259    16937 	   suppliers    TABLE     �  CREATE TABLE public.suppliers (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    categories character varying(255)[],
    phone character varying(50),
    email character varying(255),
    address character varying(255),
    status character varying(20) DEFAULT 'active'::character varying,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
    DROP TABLE public.suppliers;
       public         heap r       postgres    false            �            1259    16936    suppliers_id_seq    SEQUENCE     �   CREATE SEQUENCE public.suppliers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 '   DROP SEQUENCE public.suppliers_id_seq;
       public               postgres    false    226            |           0    0    suppliers_id_seq    SEQUENCE OWNED BY     E   ALTER SEQUENCE public.suppliers_id_seq OWNED BY public.suppliers.id;
          public               postgres    false    225            �            1259    16876    tables    TABLE       CREATE TABLE public.tables (
    id integer NOT NULL,
    number integer NOT NULL,
    seats integer NOT NULL,
    status character varying(20) DEFAULT 'free'::character varying NOT NULL,
    reserved_at timestamp without time zone,
    occupied_at timestamp without time zone
);
    DROP TABLE public.tables;
       public         heap r       postgres    false            �            1259    16875    tables_id_seq    SEQUENCE     �   CREATE SEQUENCE public.tables_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 $   DROP SEQUENCE public.tables_id_seq;
       public               postgres    false    220            }           0    0    tables_id_seq    SEQUENCE OWNED BY     ?   ALTER SEQUENCE public.tables_id_seq OWNED BY public.tables.id;
          public               postgres    false    219            �            1259    16841    users    TABLE     D  CREATE TABLE public.users (
    id bigint NOT NULL,
    username text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    role text,
    status text,
    last_active timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name character varying(255)
);
    DROP TABLE public.users;
       public         heap r       postgres    false            �            1259    16840    users_id_seq    SEQUENCE     u   CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 #   DROP SEQUENCE public.users_id_seq;
       public               postgres    false    218            ~           0    0    users_id_seq    SEQUENCE OWNED BY     =   ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;
          public               postgres    false    217            u           2604    17117    categories id    DEFAULT     n   ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);
 <   ALTER TABLE public.categories ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    232    231    232            x           2604    17128 	   dishes id    DEFAULT     f   ALTER TABLE ONLY public.dishes ALTER COLUMN id SET DEFAULT nextval('public.dishes_id_seq'::regclass);
 8   ALTER TABLE public.dishes ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    234    233    234            i           2604    16931    inventory id    DEFAULT     l   ALTER TABLE ONLY public.inventory ALTER COLUMN id SET DEFAULT nextval('public.inventory_id_seq'::regclass);
 ;   ALTER TABLE public.inventory ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    224    223    224            }           2604    17145    menu_item_options id    DEFAULT     |   ALTER TABLE ONLY public.menu_item_options ALTER COLUMN id SET DEFAULT nextval('public.menu_item_options_id_seq'::regclass);
 C   ALTER TABLE public.menu_item_options ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    236    235    236            �           2604    17180    order_item_options id    DEFAULT     ~   ALTER TABLE ONLY public.order_item_options ALTER COLUMN id SET DEFAULT nextval('public.order_item_options_id_seq'::regclass);
 D   ALTER TABLE public.order_item_options ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    240    239    240            �           2604    17159    order_items id    DEFAULT     p   ALTER TABLE ONLY public.order_items ALTER COLUMN id SET DEFAULT nextval('public.order_items_id_seq'::regclass);
 =   ALTER TABLE public.order_items ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    237    238    238            e           2604    16891 	   orders id    DEFAULT     f   ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);
 8   ALTER TABLE public.orders ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    222    221    222            t           2604    16969    request_items id    DEFAULT     t   ALTER TABLE ONLY public.request_items ALTER COLUMN id SET DEFAULT nextval('public.request_items_id_seq'::regclass);
 ?   ALTER TABLE public.request_items ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    229    230    230            p           2604    16952    requests id    DEFAULT     j   ALTER TABLE ONLY public.requests ALTER COLUMN id SET DEFAULT nextval('public.requests_id_seq'::regclass);
 :   ALTER TABLE public.requests ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    227    228    228            �           2604    17404    shift_employees id    DEFAULT     x   ALTER TABLE ONLY public.shift_employees ALTER COLUMN id SET DEFAULT nextval('public.shift_employees_id_seq'::regclass);
 A   ALTER TABLE public.shift_employees ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    244    243    244            �           2604    17388 	   shifts id    DEFAULT     f   ALTER TABLE ONLY public.shifts ALTER COLUMN id SET DEFAULT nextval('public.shifts_id_seq'::regclass);
 8   ALTER TABLE public.shifts ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    241    242    242            l           2604    16940    suppliers id    DEFAULT     l   ALTER TABLE ONLY public.suppliers ALTER COLUMN id SET DEFAULT nextval('public.suppliers_id_seq'::regclass);
 ;   ALTER TABLE public.suppliers ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    226    225    226            c           2604    16879 	   tables id    DEFAULT     f   ALTER TABLE ONLY public.tables ALTER COLUMN id SET DEFAULT nextval('public.tables_id_seq'::regclass);
 8   ALTER TABLE public.tables ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    219    220    220            b           2604    16844    users id    DEFAULT     d   ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
 7   ALTER TABLE public.users ALTER COLUMN id DROP DEFAULT;
       public               postgres    false    217    218    218            ^          0    17114 
   categories 
   TABLE DATA           F   COPY public.categories (id, name, created_at, updated_at) FROM stdin;
    public               postgres    false    232   ��       `          0    17125    dishes 
   TABLE DATA           �   COPY public.dishes (id, category_id, name, price, image_url, is_available, preparation_time, calories, allergens, created_at, updated_at) FROM stdin;
    public               postgres    false    234   ?�       V          0    16928 	   inventory 
   TABLE DATA           u   COPY public.inventory (id, name, category, quantity, unit, min_quantity, branch, created_at, updated_at) FROM stdin;
    public               postgres    false    224   t�       b          0    17142    menu_item_options 
   TABLE DATA           \   COPY public.menu_item_options (id, dish_id, name, price_modifier, is_available) FROM stdin;
    public               postgres    false    236   ˞       f          0    17177    order_item_options 
   TABLE DATA           J   COPY public.order_item_options (id, order_item_id, option_id) FROM stdin;
    public               postgres    false    240   �       d          0    17156    order_items 
   TABLE DATA           l   COPY public.order_items (id, order_id, dish_id, quantity, price, notes, created_at, updated_at) FROM stdin;
    public               postgres    false    238   �       T          0    16888    orders 
   TABLE DATA           �   COPY public.orders (id, table_id, waiter_id, status, total_amount, comment, created_at, updated_at, completed_at, cancelled_at) FROM stdin;
    public               postgres    false    222   ��       \          0    16966    request_items 
   TABLE DATA           O   COPY public.request_items (id, request_id, inventory_id, quantity) FROM stdin;
    public               postgres    false    230   ��       Z          0    16949    requests 
   TABLE DATA           w   COPY public.requests (id, branch, supplier_id, items, priority, comment, status, created_at, completed_at) FROM stdin;
    public               postgres    false    228   ��       j          0    17401    shift_employees 
   TABLE DATA           P   COPY public.shift_employees (id, shift_id, employee_id, created_at) FROM stdin;
    public               postgres    false    244   �       h          0    17385    shifts 
   TABLE DATA           k   COPY public.shifts (id, date, start_time, end_time, manager_id, notes, created_at, updated_at) FROM stdin;
    public               postgres    false    242   ڡ       X          0    16937 	   suppliers 
   TABLE DATA           p   COPY public.suppliers (id, name, categories, phone, email, address, status, created_at, updated_at) FROM stdin;
    public               postgres    false    226   /�       R          0    16876    tables 
   TABLE DATA           U   COPY public.tables (id, number, seats, status, reserved_at, occupied_at) FROM stdin;
    public               postgres    false    220   �       P          0    16841    users 
   TABLE DATA           w   COPY public.users (id, username, email, password, role, status, last_active, created_at, updated_at, name) FROM stdin;
    public               postgres    false    218   ~�                  0    0    categories_id_seq    SEQUENCE SET     ?   SELECT pg_catalog.setval('public.categories_id_seq', 5, true);
          public               postgres    false    231            �           0    0    dishes_id_seq    SEQUENCE SET     <   SELECT pg_catalog.setval('public.dishes_id_seq', 20, true);
          public               postgres    false    233            �           0    0    inventory_id_seq    SEQUENCE SET     >   SELECT pg_catalog.setval('public.inventory_id_seq', 9, true);
          public               postgres    false    223            �           0    0    menu_item_options_id_seq    SEQUENCE SET     G   SELECT pg_catalog.setval('public.menu_item_options_id_seq', 1, false);
          public               postgres    false    235            �           0    0    order_item_options_id_seq    SEQUENCE SET     H   SELECT pg_catalog.setval('public.order_item_options_id_seq', 1, false);
          public               postgres    false    239            �           0    0    order_items_id_seq    SEQUENCE SET     @   SELECT pg_catalog.setval('public.order_items_id_seq', 8, true);
          public               postgres    false    237            �           0    0    orders_id_seq    SEQUENCE SET     ;   SELECT pg_catalog.setval('public.orders_id_seq', 8, true);
          public               postgres    false    221            �           0    0    request_items_id_seq    SEQUENCE SET     C   SELECT pg_catalog.setval('public.request_items_id_seq', 1, false);
          public               postgres    false    229            �           0    0    requests_id_seq    SEQUENCE SET     =   SELECT pg_catalog.setval('public.requests_id_seq', 4, true);
          public               postgres    false    227            �           0    0    shift_employees_id_seq    SEQUENCE SET     E   SELECT pg_catalog.setval('public.shift_employees_id_seq', 60, true);
          public               postgres    false    243            �           0    0    shifts_id_seq    SEQUENCE SET     ;   SELECT pg_catalog.setval('public.shifts_id_seq', 1, true);
          public               postgres    false    241            �           0    0    suppliers_id_seq    SEQUENCE SET     >   SELECT pg_catalog.setval('public.suppliers_id_seq', 3, true);
          public               postgres    false    225            �           0    0    tables_id_seq    SEQUENCE SET     ;   SELECT pg_catalog.setval('public.tables_id_seq', 6, true);
          public               postgres    false    219            �           0    0    users_id_seq    SEQUENCE SET     ;   SELECT pg_catalog.setval('public.users_id_seq', 41, true);
          public               postgres    false    217            �           2606    17123    categories categories_pkey 
   CONSTRAINT     X   ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);
 D   ALTER TABLE ONLY public.categories DROP CONSTRAINT categories_pkey;
       public                 postgres    false    232            �           2606    17135    dishes dishes_pkey 
   CONSTRAINT     P   ALTER TABLE ONLY public.dishes
    ADD CONSTRAINT dishes_pkey PRIMARY KEY (id);
 <   ALTER TABLE ONLY public.dishes DROP CONSTRAINT dishes_pkey;
       public                 postgres    false    234            �           2606    16935    inventory inventory_pkey 
   CONSTRAINT     V   ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_pkey PRIMARY KEY (id);
 B   ALTER TABLE ONLY public.inventory DROP CONSTRAINT inventory_pkey;
       public                 postgres    false    224            �           2606    17149 (   menu_item_options menu_item_options_pkey 
   CONSTRAINT     f   ALTER TABLE ONLY public.menu_item_options
    ADD CONSTRAINT menu_item_options_pkey PRIMARY KEY (id);
 R   ALTER TABLE ONLY public.menu_item_options DROP CONSTRAINT menu_item_options_pkey;
       public                 postgres    false    236            �           2606    17182 *   order_item_options order_item_options_pkey 
   CONSTRAINT     h   ALTER TABLE ONLY public.order_item_options
    ADD CONSTRAINT order_item_options_pkey PRIMARY KEY (id);
 T   ALTER TABLE ONLY public.order_item_options DROP CONSTRAINT order_item_options_pkey;
       public                 postgres    false    240            �           2606    17165    order_items order_items_pkey 
   CONSTRAINT     Z   ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);
 F   ALTER TABLE ONLY public.order_items DROP CONSTRAINT order_items_pkey;
       public                 postgres    false    238            �           2606    16898    orders orders_pkey 
   CONSTRAINT     P   ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);
 <   ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_pkey;
       public                 postgres    false    222            �           2606    16971     request_items request_items_pkey 
   CONSTRAINT     ^   ALTER TABLE ONLY public.request_items
    ADD CONSTRAINT request_items_pkey PRIMARY KEY (id);
 J   ALTER TABLE ONLY public.request_items DROP CONSTRAINT request_items_pkey;
       public                 postgres    false    230            �           2606    16959    requests requests_pkey 
   CONSTRAINT     T   ALTER TABLE ONLY public.requests
    ADD CONSTRAINT requests_pkey PRIMARY KEY (id);
 @   ALTER TABLE ONLY public.requests DROP CONSTRAINT requests_pkey;
       public                 postgres    false    228            �           2606    17407 $   shift_employees shift_employees_pkey 
   CONSTRAINT     b   ALTER TABLE ONLY public.shift_employees
    ADD CONSTRAINT shift_employees_pkey PRIMARY KEY (id);
 N   ALTER TABLE ONLY public.shift_employees DROP CONSTRAINT shift_employees_pkey;
       public                 postgres    false    244            �           2606    17409 8   shift_employees shift_employees_shift_id_employee_id_key 
   CONSTRAINT     �   ALTER TABLE ONLY public.shift_employees
    ADD CONSTRAINT shift_employees_shift_id_employee_id_key UNIQUE (shift_id, employee_id);
 b   ALTER TABLE ONLY public.shift_employees DROP CONSTRAINT shift_employees_shift_id_employee_id_key;
       public                 postgres    false    244    244            �           2606    17394    shifts shifts_pkey 
   CONSTRAINT     P   ALTER TABLE ONLY public.shifts
    ADD CONSTRAINT shifts_pkey PRIMARY KEY (id);
 <   ALTER TABLE ONLY public.shifts DROP CONSTRAINT shifts_pkey;
       public                 postgres    false    242            �           2606    16947    suppliers suppliers_pkey 
   CONSTRAINT     V   ALTER TABLE ONLY public.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);
 B   ALTER TABLE ONLY public.suppliers DROP CONSTRAINT suppliers_pkey;
       public                 postgres    false    226            �           2606    16886    tables tables_number_key 
   CONSTRAINT     U   ALTER TABLE ONLY public.tables
    ADD CONSTRAINT tables_number_key UNIQUE (number);
 B   ALTER TABLE ONLY public.tables DROP CONSTRAINT tables_number_key;
       public                 postgres    false    220            �           2606    16884    tables tables_pkey 
   CONSTRAINT     P   ALTER TABLE ONLY public.tables
    ADD CONSTRAINT tables_pkey PRIMARY KEY (id);
 <   ALTER TABLE ONLY public.tables DROP CONSTRAINT tables_pkey;
       public                 postgres    false    220            �           2606    16850    users users_pkey 
   CONSTRAINT     N   ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
 :   ALTER TABLE ONLY public.users DROP CONSTRAINT users_pkey;
       public                 postgres    false    218            �           1259    17193    idx_dishes_category    INDEX     M   CREATE INDEX idx_dishes_category ON public.dishes USING btree (category_id);
 '   DROP INDEX public.idx_dishes_category;
       public                 postgres    false    234            �           1259    17194    idx_menu_item_options_dish    INDEX     [   CREATE INDEX idx_menu_item_options_dish ON public.menu_item_options USING btree (dish_id);
 .   DROP INDEX public.idx_menu_item_options_dish;
       public                 postgres    false    236            �           1259    17197 !   idx_order_item_options_order_item    INDEX     i   CREATE INDEX idx_order_item_options_order_item ON public.order_item_options USING btree (order_item_id);
 5   DROP INDEX public.idx_order_item_options_order_item;
       public                 postgres    false    240            �           1259    17196    idx_order_items_dish    INDEX     O   CREATE INDEX idx_order_items_dish ON public.order_items USING btree (dish_id);
 (   DROP INDEX public.idx_order_items_dish;
       public                 postgres    false    238            �           1259    17195    idx_order_items_order    INDEX     Q   CREATE INDEX idx_order_items_order ON public.order_items USING btree (order_id);
 )   DROP INDEX public.idx_order_items_order;
       public                 postgres    false    238            �           1259    16851    idx_users_email    INDEX     I   CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);
 #   DROP INDEX public.idx_users_email;
       public                 postgres    false    218            �           1259    16852    idx_users_username    INDEX     O   CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);
 &   DROP INDEX public.idx_users_username;
       public                 postgres    false    218            �           2606    17136    dishes dishes_category_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.dishes
    ADD CONSTRAINT dishes_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE SET NULL;
 H   ALTER TABLE ONLY public.dishes DROP CONSTRAINT dishes_category_id_fkey;
       public               postgres    false    234    4764    232            �           2606    17150 0   menu_item_options menu_item_options_dish_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.menu_item_options
    ADD CONSTRAINT menu_item_options_dish_id_fkey FOREIGN KEY (dish_id) REFERENCES public.dishes(id) ON DELETE CASCADE;
 Z   ALTER TABLE ONLY public.menu_item_options DROP CONSTRAINT menu_item_options_dish_id_fkey;
       public               postgres    false    236    234    4766            �           2606    17188 4   order_item_options order_item_options_option_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.order_item_options
    ADD CONSTRAINT order_item_options_option_id_fkey FOREIGN KEY (option_id) REFERENCES public.menu_item_options(id) ON DELETE SET NULL;
 ^   ALTER TABLE ONLY public.order_item_options DROP CONSTRAINT order_item_options_option_id_fkey;
       public               postgres    false    4770    240    236            �           2606    17183 8   order_item_options order_item_options_order_item_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.order_item_options
    ADD CONSTRAINT order_item_options_order_item_id_fkey FOREIGN KEY (order_item_id) REFERENCES public.order_items(id) ON DELETE CASCADE;
 b   ALTER TABLE ONLY public.order_item_options DROP CONSTRAINT order_item_options_order_item_id_fkey;
       public               postgres    false    240    238    4774            �           2606    17171 $   order_items order_items_dish_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_dish_id_fkey FOREIGN KEY (dish_id) REFERENCES public.dishes(id) ON DELETE SET NULL;
 N   ALTER TABLE ONLY public.order_items DROP CONSTRAINT order_items_dish_id_fkey;
       public               postgres    false    4766    238    234            �           2606    17166 %   order_items order_items_order_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;
 O   ALTER TABLE ONLY public.order_items DROP CONSTRAINT order_items_order_id_fkey;
       public               postgres    false    238    4754    222            �           2606    16899    orders orders_table_id_fkey    FK CONSTRAINT     |   ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_table_id_fkey FOREIGN KEY (table_id) REFERENCES public.tables(id);
 E   ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_table_id_fkey;
       public               postgres    false    4752    222    220            �           2606    16904    orders orders_waiter_id_fkey    FK CONSTRAINT     }   ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_waiter_id_fkey FOREIGN KEY (waiter_id) REFERENCES public.users(id);
 F   ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_waiter_id_fkey;
       public               postgres    false    218    4748    222            �           2606    16977 -   request_items request_items_inventory_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.request_items
    ADD CONSTRAINT request_items_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES public.inventory(id);
 W   ALTER TABLE ONLY public.request_items DROP CONSTRAINT request_items_inventory_id_fkey;
       public               postgres    false    230    4756    224            �           2606    16972 +   request_items request_items_request_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.request_items
    ADD CONSTRAINT request_items_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.requests(id) ON DELETE CASCADE;
 U   ALTER TABLE ONLY public.request_items DROP CONSTRAINT request_items_request_id_fkey;
       public               postgres    false    4760    228    230            �           2606    16960 "   requests requests_supplier_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.requests
    ADD CONSTRAINT requests_supplier_id_fkey FOREIGN KEY (supplier_id) REFERENCES public.suppliers(id);
 L   ALTER TABLE ONLY public.requests DROP CONSTRAINT requests_supplier_id_fkey;
       public               postgres    false    228    4758    226            �           2606    17415 0   shift_employees shift_employees_employee_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.shift_employees
    ADD CONSTRAINT shift_employees_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.users(id);
 Z   ALTER TABLE ONLY public.shift_employees DROP CONSTRAINT shift_employees_employee_id_fkey;
       public               postgres    false    4748    244    218            �           2606    17410 -   shift_employees shift_employees_shift_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.shift_employees
    ADD CONSTRAINT shift_employees_shift_id_fkey FOREIGN KEY (shift_id) REFERENCES public.shifts(id) ON DELETE CASCADE;
 W   ALTER TABLE ONLY public.shift_employees DROP CONSTRAINT shift_employees_shift_id_fkey;
       public               postgres    false    4779    242    244            �           2606    17395    shifts shifts_manager_id_fkey    FK CONSTRAINT        ALTER TABLE ONLY public.shifts
    ADD CONSTRAINT shifts_manager_id_fkey FOREIGN KEY (manager_id) REFERENCES public.users(id);
 G   ALTER TABLE ONLY public.shifts DROP CONSTRAINT shifts_manager_id_fkey;
       public               postgres    false    242    4748    218            ^   v   x����	�0@�s2�(�7��,S졇B{�7HA!(�޿Q��C�>�����/��X
�l�d&qJ�Z
�{iz�.x��'�6���F�Q��?��g��9.v����Lf���BU-      `   %  x��R�n�@��_q?`���a���Р��h�6Hy���&J��1HN� ���='u���Ռfv�����
�(9��B����A�Y�u�5�=-n��������*kf�͌I�Y*��"�҈��^���L��<�;z�>q&��vޠ��Q��-j4bJt�ٚ�4٘,�UP^�T8���(�.ë$�E�����v�WwK
����'	ϗ�o��QR%�]���N�EV?��hɄ�7�f~����T>ٹ<:��_$�`��gp?�]�@'bϱ�s�+�,�"���]�&x      V   G  x���;n1���)� ����^�%J�C��� ��Q�ti	A�0�Q�� A����H����36��\_��Y��6=�F�d-Y�PW�z͕C Fvmtm��h(Z�}�U�HY����4~Eж<���&:�!�+ � ���"Sy�k�O�i!F�{��Ek
H���)w��M�:Mdޒ�t�\$��2ݥ�_����B� y2��#p#�]�B���c�\�.�^��F2?��;�2G�#���L8OAV��G�����ύ��98�\�LS�R|����Vr�侌6�j_�����+�6�[X�51�RZ)�	�s      b      x������ � �      f      x������ � �      d   �   x����	1EѵT�40B_[r-鿎��0��/�< `�L� �7�-�R[�K�T�po*x��'��&�'�#����c�<�5ťI�?����3/�徂�2��I8��|�O΄	bg�|vb	Sp�j�0!��5��l�~�$�қ�/B�71�g�      T   �   x�}�INAE��S��<��Cp�l��V����T%����ٿʁA^�>���{}�D B�{�=��h(�9�����Qh�ػ��#txZ*بrT�C���N����%TP*��#��:����	{s�u�������*����%��z/c��1�J�TJ�^�j;������*wT�P��S����,��g�%����p\�sX�eRk��/1V>�k��e�Xwp      \      x������ � �      Z   �   x���;
�@���fCj��)�wi�͈�X(�l-]DH�����D�	�X��o>��I�$!3c6�|ڣ=�3v�a��
�mՇ�����"��I�*�8�ʣ�c��PK��e�"��#��5>����u����f�Y��2b`B)Y40�_��dB�R�Џ%��N�;����xr      j   K   x�}���0C�s��$2'��t�9Z(��O�\M�k3;�5.�;��A��L��-�$Q��!Td!_NE�!"/@��      h   E   x�M���@��d
���K�,�?P �\X����� ����}Co��l������@��|
3_��      X   �   x���-�@F��)�j���t�m�`1J�!!@R�7@Є� W���AV�d��}�a�eX�%���v�t\E'�������q�CD��Y'�OgJ�pTNc��6]i�Jʤ`Sc�����q��s^|��[ӓ^���
����R a��� ߽!���GW�p55��d�ݖnn���(���$-�����{�      R   j   x�u���0Dk{
 �	�L@A�!�?�K�tջwg`��߭�2{0A���ƹ��u����h#���
U���dTW�_�v�����s�RU�\p	����#�      P   �  x����r�@�5>E�M���m���A�ʦD$x�E���k�CM�hp&N�$Mw�W��8� THY��[t����U,�cv��}�mvE+*��(Sf�H .��8)�{��ҡ:�l�}���?ktl�����$���s[��H��H�D#2�*A2���s�@$!fK��烞�����f���l���4+w��q+��mcB��)�A��a��.��7�i��bK%�R A(Q���*=I�sM,`eI��.���m�=��?XYe/Z��
�yT�	��ޢ(��Jvz-�S�O3+R���~�wHѰ�!�BB��7�Xy�X~�����E�z֭��'�/� L@�J|�\���7��0J�S߈8oZ��_��X�D�*P��*�7�R��*B��z�.����P�Ń�`��꺏���N��t}�[�Q�)�V�ݱ&p�[�0���>���������4���퓉�]F`Ƕ{�1_.����ٶɢ8��|�84�����X��ͬfMV^���7�d���0+��vb�a1��PF_qo�����B��zz�;zkh���g�7��Iزg�lR���y� !/�ԃp��/��
�#�4PI�]��F�������?
���8\�{�������	��\���ݦ57hic�h��v��L��.n"T�n������޹�?`r�r���@TEE���Gi>�gP(� bt�     