<?php

// 设置缓存目录（存放 Token 和过期时间等）
$tmpDir = './live_conf';

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


// 检查文件夹是否存在
if (!is_dir($tmpDir) && !file_exists($tmpDir . '/khmusic_token.txt') && !file_exists($tmpDir . '/khmusic_expiration.txt')) {
    // 若文件夹不存在，则先建立一个
    mkdir($tmpDir, 0755, true);
    // 获取 M3U8 地址
    $m3u8Link = parseKhmusic();
    // 内容写入到文件
    $liveToken = parseUrl($m3u8Link, 'token');
    $liveExpiration = parseUrl($m3u8Link, 'expiration');
    saveData('khmusic', $liveToken, $liveExpiration, $tmpDir);
} else {
    // 检查 Token 是否过期，未过期则直接读取原有的 Token
    if (!validTime(readData($tmpDir, 'khmusic', 'expiration'))) {
        // 重新获取 M3U8 地址
        $m3u8Link = parseKhmusic();
        // 内容写入到文件
        $liveToken = parseUrl($m3u8Link, 'token');
        $liveExpiration = parseUrl($m3u8Link, 'expiration');
        saveData('khmusic', $liveToken, $liveExpiration, $tmpDir);
    } else {
        // 从文件读取原有 Token
        $liveToken = readData($tmpDir, 'khmusic', 'token');
        $m3u8Link = 'https://vohradiow-hichannel.cdn.hinet.net/live/RA000077/chunklist.m3u8?token=' . $liveToken . '&expires=' . readData($tmpDir, 'khmusic', 'expiration');
    }
}

// 替换 TS 分片文件名
$streamContents = file_get_contents($m3u8Link);

// 返回 M3U8 内容
header('Content-Type: application/vnd.apple.mpegurl');
echo str_replace('media_', 'https://vohradiow-hichannel.cdn.hinet.net/live/RA000077/media_', $streamContents);

?>
