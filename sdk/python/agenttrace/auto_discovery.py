"""
Auto-discovery module for detecting AI frameworks at import time.

Detects installed frameworks (LangChain, CrewAI, OpenAI, Anthropic, LlamaIndex, AutoGen)
and enables zero-configuration instrumentation.
"""

import importlib
import importlib.metadata
import logging
import os
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional

logger = logging.getLogger("agenttrace.auto_discovery")


class FrameworkType(str, Enum):
    OPENAI = "openai"
    ANTHROPIC = "anthropic"
    LANGCHAIN = "langchain"
    LLAMAINDEX = "llamaindex"
    CREWAI = "crewai"
    AUTOGEN = "autogen"
    UNKNOWN = "unknown"


class DiscoveryStatus(str, Enum):
    NOT_FOUND = "not_found"
    DETECTED = "detected"
    INSTRUMENTED = "instrumented"
    FAILED = "failed"
    DISABLED = "disabled"


@dataclass
class DiscoveredFramework:
    framework: FrameworkType
    status: DiscoveryStatus
    version: Optional[str] = None
    package_name: Optional[str] = None
    error: Optional[str] = None


@dataclass
class DiscoveryResult:
    frameworks: List[DiscoveredFramework] = field(default_factory=list)
    auto_instrumented: List[FrameworkType] = field(default_factory=list)
    errors: List[str] = field(default_factory=list)


# Mapping of framework types to their detection metadata
FRAMEWORK_REGISTRY: Dict[FrameworkType, dict] = {
    FrameworkType.OPENAI: {
        "packages": ["openai"],
        "import_check": "openai",
        "instrumentation": "agenttrace.integrations.openai",
        "enable_fn": "OpenAIInstrumentation.enable",
    },
    FrameworkType.ANTHROPIC: {
        "packages": ["anthropic"],
        "import_check": "anthropic",
        "instrumentation": "agenttrace.integrations.anthropic",
        "enable_fn": "AnthropicInstrumentation.enable",
    },
    FrameworkType.LANGCHAIN: {
        "packages": ["langchain", "langchain-core", "langchain-community"],
        "import_check": "langchain",
        "instrumentation": "agenttrace.integrations.langchain",
        "enable_fn": "LangChainInstrumentation.enable",
    },
    FrameworkType.LLAMAINDEX: {
        "packages": ["llama-index", "llama-index-core"],
        "import_check": "llama_index",
        "instrumentation": "agenttrace.integrations.llamaindex",
        "enable_fn": "LlamaIndexInstrumentation.enable",
    },
    FrameworkType.CREWAI: {
        "packages": ["crewai"],
        "import_check": "crewai",
        "instrumentation": None,
        "enable_fn": None,
    },
    FrameworkType.AUTOGEN: {
        "packages": ["pyautogen", "autogen"],
        "import_check": "autogen",
        "instrumentation": None,
        "enable_fn": None,
    },
}


def _get_package_version(package_names: List[str]) -> Optional[str]:
    """Try to get the version of a package from installed metadata."""
    for name in package_names:
        try:
            return importlib.metadata.version(name)
        except importlib.metadata.PackageNotFoundError:
            continue
    return None


def _is_package_importable(import_name: str) -> bool:
    """Check if a package can be imported without actually importing it."""
    try:
        spec = importlib.util.find_spec(import_name)
        return spec is not None
    except (ModuleNotFoundError, ValueError):
        return False


def detect_frameworks() -> DiscoveryResult:
    """
    Detect installed AI frameworks using importlib.metadata and import checks.
    Returns a DiscoveryResult with all detected frameworks.
    """
    result = DiscoveryResult()

    for framework_type, meta in FRAMEWORK_REGISTRY.items():
        version = _get_package_version(meta["packages"])
        importable = _is_package_importable(meta["import_check"])

        if version or importable:
            result.frameworks.append(
                DiscoveredFramework(
                    framework=framework_type,
                    status=DiscoveryStatus.DETECTED,
                    version=version,
                    package_name=meta["packages"][0],
                )
            )
        else:
            result.frameworks.append(
                DiscoveredFramework(
                    framework=framework_type,
                    status=DiscoveryStatus.NOT_FOUND,
                    package_name=meta["packages"][0],
                )
            )

    return result


def auto_instrument(
    frameworks: Optional[List[FrameworkType]] = None,
    exclude: Optional[List[FrameworkType]] = None,
) -> DiscoveryResult:
    """
    Auto-detect and instrument all discovered frameworks.

    Args:
        frameworks: If specified, only instrument these frameworks. Otherwise detect all.
        exclude: Frameworks to skip even if detected.
    """
    if os.environ.get("AGENTTRACE_DISABLE_AUTO_INSTRUMENT", "").lower() in ("1", "true"):
        logger.info("Auto-instrumentation disabled via AGENTTRACE_DISABLE_AUTO_INSTRUMENT")
        return DiscoveryResult()

    result = detect_frameworks()
    exclude_set = set(exclude or [])

    for discovered in result.frameworks:
        if discovered.status != DiscoveryStatus.DETECTED:
            continue

        if frameworks and discovered.framework not in frameworks:
            continue

        if discovered.framework in exclude_set:
            discovered.status = DiscoveryStatus.DISABLED
            continue

        meta = FRAMEWORK_REGISTRY.get(discovered.framework, {})
        instrumentation_module = meta.get("instrumentation")

        if not instrumentation_module:
            logger.debug(
                "No instrumentation available for %s", discovered.framework.value
            )
            continue

        try:
            mod = importlib.import_module(instrumentation_module)
            enable_fn_path = meta.get("enable_fn", "")
            if enable_fn_path:
                parts = enable_fn_path.split(".")
                obj = mod
                for part in parts:
                    obj = getattr(obj, part)
                if callable(obj):
                    obj()

            discovered.status = DiscoveryStatus.INSTRUMENTED
            result.auto_instrumented.append(discovered.framework)
            logger.info(
                "Auto-instrumented %s (v%s)",
                discovered.framework.value,
                discovered.version or "unknown",
            )
        except Exception as e:
            discovered.status = DiscoveryStatus.FAILED
            discovered.error = str(e)
            result.errors.append(
                f"Failed to instrument {discovered.framework.value}: {e}"
            )
            logger.warning(
                "Failed to auto-instrument %s: %s", discovered.framework.value, e
            )

    return result


def print_discovery_report(result: DiscoveryResult) -> None:
    """Print a human-readable discovery report to stdout."""
    print("\n🔍 AgentTrace Auto-Discovery Report")
    print("=" * 50)

    for fw in result.frameworks:
        status_icon = {
            DiscoveryStatus.NOT_FOUND: "⬜",
            DiscoveryStatus.DETECTED: "🔵",
            DiscoveryStatus.INSTRUMENTED: "✅",
            DiscoveryStatus.FAILED: "❌",
            DiscoveryStatus.DISABLED: "⏭️",
        }.get(fw.status, "❓")

        version_str = f" v{fw.version}" if fw.version else ""
        error_str = f" ({fw.error})" if fw.error else ""
        print(f"  {status_icon} {fw.framework.value}{version_str} - {fw.status.value}{error_str}")

    if result.auto_instrumented:
        print(f"\n✨ Auto-instrumented: {', '.join(f.value for f in result.auto_instrumented)}")
    else:
        print("\nℹ️  No frameworks were auto-instrumented.")

    if result.errors:
        print(f"\n⚠️  {len(result.errors)} error(s) during instrumentation")

    print()
