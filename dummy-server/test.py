import json
import unittest

import ddt
from main import fix_json


@ddt.ddt()
class MyTestCase(unittest.TestCase):
    @ddt.data(
        [
            b"""{"token":"fakeToken", events: [{attrs:{message:`s\x00\x00\x00\neventThree},ts:"3"}], logs: [], threads: [], client_time: 1 }""",  # noqa E501
            b"""{"token":"fakeToken", "events": [{"attrs":{"message":"eventThree"},"ts":"3"}], "logs": [], "threads": [], "client_time": 1 }""",  # noqa E501
        ],
        [
            b"""{"token":"fakeToken", events: [{attrs:{message:`s\x00\x00\x00\x08eventOne},ts:"1"}], logs: [], threads: [], client_time: 1 }""",  # noqa E501
            b"""{"token":"fakeToken", "events": [{"attrs":{"message":"eventOne"},"ts":"1"}], "logs": [], "threads": [], "client_time": 1 }""",  # noqa E501
        ],
        [
            b"""{"token":"fakeToken", events: [{attrs:{message:`s\x00\x00\x00\x08eventOne},ts:"1"},{attrs:{message:`s\x00\x00\x00\x08eventTwo},ts:"1"}], logs: [], threads: [], client_time: 1 }""",  # noqa E501
            b"""{"token":"fakeToken", "events": [{"attrs":{"message":"eventOne"},"ts":"1"},{"attrs":{"message":"eventTwo"},"ts":"1"}], "logs": [], "threads": [], "client_time": 1 }""",  # noqa E501
        ],
        [
            b"""{"token":"fakeToken", events: [{attrs:{message:`s\x00\x00\x00\x08event"ne},ts:"1"}], logs: [], threads: [], client_time: 1 }""",  # noqa E501
            b"""{"token":"fakeToken", "events": [{"attrs":{"message":"event\\"ne"},"ts":"1"}], "logs": [], "threads": [], "client_time": 1 }""",  # noqa E501
        ],
        [
            b"""{"token":"fakeToken", events: [{attrs:{message:`s\x00\x00\x00\x08eventne\n},ts:"1"}], logs: [], threads: [], client_time: 1 }""",  # noqa E501
            b"""{"token":"fakeToken", "events": [{"attrs":{"message":"eventne "},"ts":"1"}], "logs": [], "threads": [], "client_time": 1 }""",  # noqa E501
        ],
        [
            b"""{"token":"fakeToken", events: [{attrs:{message:`s\x00\x00\x00\x08eventne\\},ts:"1"}], logs: [], threads: [], client_time: 1 }""",  # noqa E501
            b"""{"token":"fakeToken", "events": [{"attrs":{"message":"eventne "},"ts":"1"}], "logs": [], "threads": [], "client_time": 1 }""",  # noqa E501
        ],
        # [
        #     b'{"token":"0JZ656PoUOqTkxqNETMzlXZY5FlliVdgxIiUuGe6skb0-","session":"ul9cOH-6D_7ZGnmGF3mzoanz5uc=","threads":[],"sessionInfo":{"serverHost":"77bc751aeac1"}, events: [{thread:"log_1", log:"log_1", attrs:{"parser":"scalyrAgentLog","logfile":"/var/log/scalyr-agent-2/agent.log",message:`s\x00\x00\x01\xb82023-03-08 10:49:22.488Z INFO [core] [scalyr-agent-2:1648] Starting scalyr agent... (version=2.1.40 revision=1284e199fa9d93a6366fd147fc73b03d8e90712a) (pid=1) (tid=281472947088864) (hostname=77bc751aeac1) (Python version=3.8.16 (default, Jan 18 2023, 02:50:12) [GCC 10.2.1 20210110]) (OpenSSL version=OpenSSL 1.1.1n  15 Mar 2022) (default fs encoding=utf-8) (locale=en_US.UTF-8) (LANG env variable=C.UTF-8) (date parsing library=udatetime)\n},ts:"1678273945400859648",si:"54605184-974b-46e5-8365-06839e286571",sn:440},{thread:"log_1", log:"log_1", attrs:{"parser":"scalyrAgentLog","logfile":"/var/log/scalyr-agent-2/agent.log",message:`s\x00\x00\x00\x952023-03-08 10:49:22.489Z INFO [core] [scalyr_agent.platform_posix:215] Emitting log lines saved during initialization: (pid=1) (tid=281472947088864)\n},sd:149,ts:"1678273945401097472"},{thread:"log_1", log:"log_1", attrs:{"parser":"scalyrAgentLog","logfile":"/var/log/scalyr-agent-2/agent.log",message:`s\x00\x00\x00\xa72023-03-08 10:49:22.489Z INFO [core] [scalyr_agent.platform_posix:221]      Checked pidfile: does not exist (pid=1) (tid=281472947088864)   (2023-03-08 10:49:22.476Z)\n},sd:167,ts:"1678273945401141504"},{thread:"log_1", log:"log_1", attrs:{"parser":"scalyrAgentLog","logfile":"/var/log/scalyr-agent-2/agent.log",message:`s\x00\x00\x00]2023-03-08 10:49:22.490Z INFO [core] [scalyr_agent.configuration:794] Configuration settings\n},sd:93,ts:"1678273945401168896"},{thread:"log_1", log:"log_1", attrs:{"parser":"scalyrAgentLog","logfile":"/var/log/scalyr-agent-2/agent.log",message:`s\x00\x00\x00g2023-03-08 10:49:22.490Z INFO [core] [scalyr_agent.configuration:801] \tverify_server_certificate: True\n},sd:103,ts:"1678273945401193472"}]}',  # noqa E501
        #     b"",
        # ],
    )
    @ddt.unpack
    def test_fix_json(self, inp, expected):
        fixed = fix_json(inp)
        # self.assertEqual(fixed, expected)
        fixed_json = json.loads(fixed)
        expected_json = json.loads(expected)
        self.assertDictEqual(fixed_json, expected_json, f"F: {fixed}; E: {expected}")


if __name__ == "__main__":
    unittest.main()
