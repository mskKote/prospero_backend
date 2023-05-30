DROP TABLE IF EXISTS public.admins CASCADE;
DROP TABLE IF EXISTS public.publishers CASCADE;
DROP TABLE IF EXISTS public.sources_rss CASCADE;

-- adminka users
CREATE TABLE public.admins
(
    user_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name     VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL
);

-- publishers {1:N} SourcesRSS
CREATE TABLE public.publishers
(
    publisher_id UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    add_date     TIMESTAMPTZ  NOT NULL DEFAULT current_timestamp,
    name         VARCHAR(100) NOT NULL UNIQUE,
    country      VARCHAR(100) NOT NULL,
    city         VARCHAR(100) NOT NULL,
    point        point        NOT NULL
);
-- RSS links and their publishers
CREATE TABLE public.sources_rss
(
    rss_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rss_url      VARCHAR(2048) UNIQUE NOT NULL,
    publisher_id UUID                 NOT NULL,
    add_date     TIMESTAMPTZ  NOT NULL DEFAULT current_timestamp,

    CONSTRAINT fk_publisher
        FOREIGN KEY (publisher_id)
            REFERENCES public.publishers (publisher_id)
);


-- test RSS
INSERT INTO public.publishers(name, country, city, point)
VALUES
    ('The New York Times', 'USA', 'New York', point(40.756133, -73.990322)),
    ('The Guardian', 'UK', 'London', point(51.534839, -0.122149)),
    ('Vedomosti', 'Russia', 'Sankt Petersburg', point(59.917904, 30.348691)),
    ('ООН', 'USA', 'New York', point(40.749571, -73.967716)),
    ('Hindustan Times', 'Delhi', 'New York', point(28.628026, 77.223106)),
    ('Rambler', 'Russia', 'Moscow', point(55.698645, 37.624570)),
    ('lenta.ru', 'Russia', 'Moscow', point(55.698645, 37.624570)),
    ('Wall Street Journal', 'USA', 'New York', point(40.749995, -73.983758)),
    ('France 24', 'France', 'Paris', point(48.830639, 2.264886)),
    ('CNN', 'US', 'Atlanta', point(33.758040, -84.394692));

INSERT INTO public.sources_rss(publisher_id, rss_url, add_date)
VALUES ((SELECT publisher_id FROM public.publishers WHERE name = 'The New York Times'), 'https://rss.nytimes.com/services/xml/rss/nyt/World.xml',  '2023-05-1 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'The Guardian'), 'https://www.theguardian.com/world/rss', '2023-05-2 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'Vedomosti'), 'https://www.vedomosti.ru/rss/news', '2023-05-3 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'ООН'), 'https://news.un.org/feed/subscribe/ru/news/all/rss.xml', '2023-05-4 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'Hindustan Times'), 'https://www.hindustantimes.com/feeds/rss/world-news/rssfeed.xml', '2023-05-5 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'Rambler'), 'https://news.rambler.ru/rss/world/', '2023-05-6 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'lenta.ru'), 'https://lenta.ru/rss/news', '2023-05-7 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'Wall Street Journal'), 'https://feeds.a.dj.com/rss/RSSWorldNews.xml', '2023-05-8 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'France 24'), 'http://america.aljazeera.com/content/ajam/articles.rss', '2023-05-9 19:30:06.887661 +00:00' :: timestamptz),
       ((SELECT publisher_id FROM public.publishers WHERE name = 'CNN'), 'http://rss.cnn.com/rss/edition_world.rss', '2023-05-10 19:30:06.887661 +00:00' :: timestamptz);

-- search

SELECT rss_id,
       rss_url,
       p.name,
       p.publisher_id,
       p.add_date,
       p.country,
       p.city,
       p.point
FROM public.sources_rss
         JOIN publishers p on p.publisher_id = sources_rss.publisher_id
WHERE LOWER(p.name) LIKE LOWER('%NEW%')
