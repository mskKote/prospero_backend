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

-- INSERT INTO public.admins(user_id, name, password)
-- VALUES (0, 'admin', '???');

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

    CONSTRAINT fk_publisher
        FOREIGN KEY (publisher_id)
            REFERENCES public.publishers (publisher_id)
);

-- test RSS
INSERT INTO public.publishers(publisher_id, name, country, city, point)
VALUES (default, 'The New York Times', 'USA', 'New York', point(40.756133, -73.990322));

INSERT INTO public.sources_rss(publisher_id, rss_url)
VALUES ((SELECT publisher_id FROM public.publishers ORDER BY add_date DESC LIMIT 1),
        'https://rss.nytimes.com/services/xml/rss/nyt/World.xml'),
       ((SELECT publisher_id FROM public.publishers ORDER BY add_date DESC LIMIT 1),
        'https://rss.nytimes.com/services/xml/rss/nyt/Technology.xml');

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
