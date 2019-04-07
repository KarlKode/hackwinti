import csv

routings = []
operations = []


def load_orders(f='production_order_headers.csv'):
    with open(f, newline='') as o_csv:
        r = csv.DictReader(o_csv)
        orders = {}
        for order in r:
            orders[order['order']] = order
        return r.fieldnames, orders


def load_materials(f='material_master.csv'):
    with open(f, newline='') as m_csv:
        r = csv.DictReader(m_csv)
        materials = {}
        for material in r:
            materials[material['material']] = material
        return r.fieldnames, materials


def load_workcenters(f='workcenters.csv'):
    with open(f, newline='') as w_csv:
        r = csv.DictReader(w_csv)
        workcenters = {}
        for workcenter in r:
            workcenters[workcenter['workcenter']] = workcenter
        return r.fieldnames, workcenters


def load_routings(f='routings.csv'):
    with open(f, newline='') as w_csv:
        r = csv.DictReader(w_csv)
        routings = {}
        for routing in r:
            k = "{}-{}/{}-{}@{}".format(routing['material'], routing['plan group'], routing['plan group counter'],
                                        routing['operation nr.'], routing['workcenter'])
            routings[k] = routing
        return r.fieldnames, routings


