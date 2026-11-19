"""Load only local extensions not integrated into SimSight itself."""
import simsight_speech as speech

_installed = False

def install():
    global _installed
    if _installed:
        return
    _installed = True
    # Scanner ranges, moving furniture, room labels, wall/surface placement
    # and layer keys now belong to the native modules and key loop.
    import importlib
    for name, explicit in (
        ('simsight_chair_alignment', False),
        ('simsight_roof_tools', False),
        ('simsight_house_planner', False),
        ('simsight_lot_grid', False),
        ('simsight_detached', False),
        ('simsight_build_coordinates', True),
        ('simsight_build_helpers', True),
        ('simsight_basement', True),
        ('simsight_house_workflow', True),
    ):
        try:
            module = importlib.import_module(name)
            if explicit:
                module.install()
        except Exception as exc:
            speech.log_line('local extension {} failed: {!r}'.format(name, exc))
    speech.log_line('local extensions installed')

