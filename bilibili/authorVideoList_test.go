package bilibili

/*
select follow, count(*)
from author
group by follow;

select a.id, a.author_name,a.follow,a.author_web_uid, vs.videos, bv.video_number
from author a
         left join (select video.author_id, count(*) as videos
                    from video
                    group by video.author_id) vs on vs.author_id = a.id
         left join bili_author_video_number bv on a.id = bv.author_id
where web_site_id = 1
  and bv.video_number is not null
  and vs.videos <> bv.video_number;



select a.author_name, video.author_id, count(video.id), bv.video_number
from video
         left join author a on a.id = video.author_id
         left join bili_author_video_number bv on video.author_id = bv.author_id
where bv.video_number is not null
group by video.author_id;
*/
