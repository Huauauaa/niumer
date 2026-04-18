import asyncio
import datetime
import requests

from playwright.async_api import async_playwright, TimeoutError


async def get_cookies():
    async with async_playwright() as p:
        browser = await p.chromium.launch(channel="chrome", headless=True)
        context = await browser.new_context()
        page = await browser.new_page()
        # TODO: login
        url = "https://example.com/login"
        await page.goto(url)
        try:
            await page.wait_for_selector(".search-total__num", state="attached", timeout=15 * 1000)
        except TimeoutError:
            print("登录失败, 请开启免密登录")
            return None
        cookies = await context.cookies()
        print("登录成功")
        return {i["name"]: i["value"] for i in cookies}


def get_id(cookies):
    # TODO: get id
    url = "https://example.com/id"
    body = {
        "rentId": "HuaWei",
        "groupId": "servicetimeflow",
        "tenantId": "",
        "url": "hr.huawei.com",
        "parentCode": "servicetimeflow",
    }
    response = requests.post(url, headers={}, json=body, cookies=cookies, verify=False)
    return response.json()["data"]["tenant"]["userAccount"]


def get_hr_id(cookies, id):
     # TODO: get hr id
    url = "https://example.com/hr-id"
    body = {
        "employeeQuery": id[1:],
        "queryDate": datetime.now().strftime("%Y-%m-%d"),
        "locale": "zh",
        "platform": "pc",
    }
    response = requests.post(url, headers={}, json=body, cookies=cookies)
    return response.json()["data"]["hrId"]


async def main():
    cookies = await get_cookies()
    if cookies is None:
        return
    id = get_id(cookies)
    hr_id = get_hr_id(cookies, id)

    # TODO
    url = "https://example.com/work-hour"
    body = {"hr_id": hr_id, "locale": "zh", "platform": "pc"}
    response = requests.post(url, headers={}, json=body, cookies=cookies)
    print(response.json())


if __name__ == "__main__":
    asyncio.run(main())
