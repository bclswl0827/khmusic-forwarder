<?php

// HiNet CDN 地址为 vohradiow-hichannel.cdn.hinet.net
// 若将来被屏蔽，可以自行搭建反代，并替换掉下面的 URL
$hinetUrl = 'https://vohradiow-hichannel.cdn.hinet.net/live';

// 设置缓存目录（存放 Token 和过期时间等）
$tmpDir = './live_cache';

// 用于保存数据
function saveData($baseName, $liveToken, $liveExpiration, $saveDir) {
    file_put_contents($saveDir . '/' . $baseName . '_token.txt', print_r($liveToken, true));
    file_put_contents($saveDir . '/' . $baseName . '_expiration.txt', print_r($liveExpiration, true));
}

// 用于读取数据
function readData($readDir, $baseName, $readType) {
    $myPath = $readDir . '/' . $baseName . '_' . $readType . '.txt';
    $myFile = fopen($myPath, 'r') or die('Unable to read file!');
    $myData = fread($myFile, filesize($myPath));
    fclose($myFile);
    return $myData;
}

// 获取直播流可用时长，并提前一小时，返回布尔值
function validTime($unixTime) {
    $currentTime = strtotime('now');
    $validHours = (strtotime($currentTime) - strtotime($unixTime)) / (60 * 60);
    // 小于 1 小时则过期，返回 false，反之返回 true
    if ($validHours < 1) {
        return true;
    } else {
        return false;
    }
}

// 透过正则解析 HLS 地址 URL 参数
function parseUrl($liveUrl, $urlParam) {
    if ($urlParam == 'token') {
        preg_match_all('/(n=).*(&)/i', $liveUrl, $regExp);
        return str_replace('&', '', str_replace('n=', '', $regExp[0][0]));
    } else {
        preg_match_all('/(s=)\d+\d$/i', $liveUrl, $regExp);
        return str_replace('s=', '', $regExp[0][0]);
    }
}

// 解析光华之声网站获取 HLS 地址
function parseKhmusic() {
    // 获取光华之声网页，并用正则提取地址
    $khmusicUrl = 'https://audio.voh.com.tw/KwongWah/m3u8.aspx';
    $webContent = str_replace('&amp;', '&', file_get_contents($khmusicUrl));
    preg_match_all('/https:\/\/vohradiow-hichannel.cdn.hinet.net\/live\/RA000077\/playlist.m3u8\?token=.+&expires=\d+/i', $webContent, $regExp);
    // 第一个即为直播地址，并替换 playlist.m3u8 为 chunklist.m3u8
    return str_replace('playlist', 'chunklist', $regExp[0][0]);
}

// 汉声电台 API 获取 HLS 地址
function parseVoh($stationNum) {
    // 构造请求体参数
    $postData = array('http' =>
        array(
            'method' => 'POST',
            'header' => 'Content-type: application/x-www-form-urlencoded; charset=UTF-8',
            'content' => http_build_query(
                array(
                    'Type' => 'VideoDisplay',
                    'PlayChannel' => $stationNum
                )
            )
        )
    );

    $apiRes = file_get_contents(
        'https://audio.voh.com.tw/API/ugC_ProgramHandle.ashx',
        false,
        stream_context_create($postData)
    );
    $jsonArray = json_decode($apiRes, true);
    return $jsonArray['Url'] . '?token=' . $jsonArray['token'] . '&expires=' . $jsonArray['expires'];
}

// 获取 URL 传递的参数
switch ($_GET['station']) {
    case 'khmusic':
        $userSelect = 'khmusic';
        $baseUrl = $hinetUrl . '/RA000077';
        break;
    case 'voh_fm':
        $userSelect = 'voh_fm';
        $vohId = '1';
        $baseUrl = $hinetUrl . '/RA000076';
        break;
    case 'voh_am':
        $userSelect = 'voh_am';
        $vohId = '2';
        $baseUrl = $hinetUrl . '/RA000074';
        break;
    default:
        echo '<html>
                <head>
                    <title>403 Forbidden</title>
                </head>
                <body>
                    <div id="msg">
                        <h2>请选择一个电台</h2>
                        <ul>
                            <li>
                                <a href="?station=khmusic" target="_blank">光华之声</a>
                            </li>
                            <li>
                                <a href="?station=voh_fm" target="_blank">汉声 FM</a>
                            </li>
                            <li>
                                <a href="?station=voh_am" target="_blank">汉声 AM</a>
                            </li>
                        </ul>
                    </div>
                </body>
            </html>';
        throw new Exception("必须选择一个项目");
}

// 检查文件夹是否存在
if (!is_dir($tmpDir)) {
    // 若文件夹不存在，则先建立一个
    mkdir($tmpDir, 0755, true);
    // 获取 M3U8 地址
    if ($userSelect == 'khmusic') {
        $m3u8Link = parseKhmusic();
    } else {
        $m3u8Link = parseVoh($vohId);
    }
    // 内容写入到文件
    $liveToken = parseUrl($m3u8Link, 'token');
    $liveExpiration = parseUrl($m3u8Link, 'expiration');
    saveData($userSelect, $liveToken, $liveExpiration, $tmpDir);
} else {
    if (!file_exists($tmpDir . '/' . $userSelect . '_token.txt') && !file_exists($tmpDir . '/' . $userSelect . '_expiration.txt')) {
        // 获取 M3U8 地址
        if ($userSelect == 'khmusic') {
            $m3u8Link = parseKhmusic();
        } else {
            $m3u8Link = parseVoh($vohId);
        }
        // 内容写入到文件
        $liveToken = parseUrl($m3u8Link, 'token');
        $liveExpiration = parseUrl($m3u8Link, 'expiration');
        saveData($userSelect, $liveToken, $liveExpiration, $tmpDir);
    } else {
        // 检查 Token 是否过期，未过期则直接读取原有的 Token
        if (!validTime(readData($tmpDir, $userSelect, 'expiration'))) {
            // 获取 M3U8 地址
            if ($userSelect == 'khmusic') {
                $m3u8Link = parseKhmusic();
            } else {
                $m3u8Link = parseVoh($vohId);
            }
            // 内容写入到文件
            $liveToken = parseUrl($m3u8Link, 'token');
            $liveExpiration = parseUrl($m3u8Link, 'expiration');
            saveData($userSelect, $liveToken, $liveExpiration, $tmpDir);
        } else {
            // 从文件读取原有 Token
            $liveToken = readData($tmpDir, $userSelect, 'token');
            $m3u8Link = $baseUrl . '/chunklist.m3u8?token=' . $liveToken . '&expires=' . readData($tmpDir, $userSelect, 'expiration');
        }
    }

}

// 替换 TS 分片文件名
$streamContents = file_get_contents($m3u8Link);

// 返回 M3U8 内容
header('Content-Type: application/vnd.apple.mpegurl');
echo str_replace('media_', $baseUrl . '/media_', $streamContents);

?>
